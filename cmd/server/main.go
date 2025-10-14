package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/kristianrpo/document-management-microservice/docs"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/events"
	httpadapter "github.com/kristianrpo/document-management-microservice/internal/adapters/http"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	cfgpkg "github.com/kristianrpo/document-management-microservice/internal/infrastructure/config"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/messaging"
	infrapkg "github.com/kristianrpo/document-management-microservice/internal/infrastructure/repository"
)

// @title Document Management Microservice API
// @version 1.0
// @description Microservice for managing document uploads, storage, and metadata
// @description
// @description Features:
// @description - Upload documents to S3
// @description - Store metadata in DynamoDB
// @description - Automatic file deduplication based on SHA256 hash
// @description - Support for multiple file types
// @description - Health check endpoint
//
// @contact.name API Support
// @contact.email support@example.com
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /
// @schemes http https
//
// @tag.name documents
// @tag.description Document upload and management operations
//
// @tag.name health
// @tag.description Health check endpoints
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system environment variables")
	}

	config := cfgpkg.Load()

	dynamoClient, err := cfgpkg.NewDynamoDBClient(
		context.Background(),
		config.AWSAccessKey,
		config.AWSSecretKey,
		config.AWSRegion,
		config.DynamoDBEndpoint,
	)
	if err != nil {
		log.Fatalf("dynamodb init: %v", err)
	}
	log.Printf("DynamoDB client initialized (endpoint: %s)", config.DynamoDBEndpoint)
	if os.Getenv("DEBUG") == "true" {
		log.Println("DEBUG mode enabled: error details will be included in responses")
	} else {
		log.Println("Set DEBUG=true to include error details in responses during local development")
	}

	s3Client, err := cfgpkg.NewS3Client(context.Background(), *config)
	if err != nil {
		log.Fatalf("s3 init: %v", err)
	}

	documentRepository := infrapkg.NewDynamoDBDocumentRepo(dynamoClient, config.DynamoDBTable)

	var objectStorage interfaces.ObjectStorage = s3Client

	fileHasher := util.NewSHA256Hasher()
	mimeDetector := util.NewExtensionBasedDetector()

	// Initialize RabbitMQ publisher for publishing events
	var messagePublisher interfaces.MessagePublisher
	if config.RabbitMQ.URL != "" {
		rabbitPublisher, err := messaging.NewRabbitMQPublisher(config.RabbitMQ)
		if err != nil {
			log.Printf("warning: failed to initialize RabbitMQ publisher: %v", err)
		} else {
			messagePublisher = rabbitPublisher
			defer func() {
				if err := messagePublisher.Close(); err != nil {
					log.Printf("Error closing message publisher: %v", err)
				}
			}()
			log.Println("RabbitMQ publisher initialized")
		}
	}

	documentService := usecases.NewDocumentService(
		documentRepository,
		objectStorage,
		fileHasher,
		mimeDetector,
	)
	documentListService := usecases.NewDocumentListService(documentRepository)
	documentGetService := usecases.NewDocumentGetService(documentRepository)
	documentDeleteService := usecases.NewDocumentDeleteService(documentRepository, objectStorage)
	documentDeleteAllService := usecases.NewDocumentDeleteAllService(documentRepository, objectStorage)
	documentTransferService := usecases.NewDocumentTransferService(documentRepository, objectStorage, 15*time.Minute)
	
	var documentRequestAuthService *usecases.DocumentRequestAuthenticationService
	if messagePublisher != nil {
		documentRequestAuthService = usecases.NewDocumentRequestAuthenticationService(
			documentRepository,
			objectStorage,
			messagePublisher,
			config.AuthenticationRequestQueue,
			24*time.Hour,
		)
	}

	errorMapper := errors.NewErrorMapper()
	errorHandler := errors.NewErrorHandler(errorMapper)

	uploadHandler := handlers.NewDocumentUploadHandler(documentService, errorHandler)
	listHandler := handlers.NewDocumentListHandler(documentListService, errorHandler)
	getHandler := handlers.NewDocumentGetHandler(documentGetService, errorHandler)
	deleteHandler := handlers.NewDocumentDeleteHandler(documentDeleteService, errorHandler)
	deleteAllHandler := handlers.NewDocumentDeleteAllHandler(documentDeleteAllService, errorHandler)
	transferHandler := handlers.NewDocumentTransferHandler(documentTransferService, errorHandler)
	
	var requestAuthHandler *handlers.DocumentRequestAuthenticationHandler
	if documentRequestAuthService != nil {
		requestAuthHandler = handlers.NewDocumentRequestAuthenticationHandler(documentRequestAuthService, errorHandler)
	}
	
	healthHandler := handlers.NewHealthHandler()

	router := httpadapter.NewRouter(uploadHandler, listHandler, getHandler, deleteHandler, deleteAllHandler, transferHandler, requestAuthHandler, healthHandler)

	// Initialize RabbitMQ consumer for event-driven communication
	ctx := context.Background()
	var messageBroker interfaces.MessageBroker

	if config.RabbitMQ.URL != "" {
		rabbitConsumer, err := messaging.NewRabbitMQConsumer(config.RabbitMQ)
		if err != nil {
			log.Printf("warning: failed to initialize RabbitMQ consumer: %v", err)
			log.Println("continuing without message broker...")
		} else {
			messageBroker = rabbitConsumer
			defer func() {
				if err := messageBroker.Close(); err != nil {
					log.Printf("Error closing message broker: %v", err)
				}
			}()

			// Set up event handler
			userTransferHandler := events.NewUserTransferHandler(documentDeleteAllService)

			// Start consuming messages
			if err := messageBroker.Subscribe(ctx, userTransferHandler.HandleUserTransferred); err != nil {
				log.Printf("warning: failed to subscribe to queue: %v", err)
			} else {
				log.Printf("listening for events on queue: %s", config.RabbitMQ.Queue)
			}
		}
	} else {
		log.Println("RabbitMQ URL not configured, skipping message broker initialization")
	}

	server := &http.Server{
		Addr:              config.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)
	<-stopSignal

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}

	log.Println("server stopped")
}
