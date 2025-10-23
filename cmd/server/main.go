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
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/middleware"
	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	cfgpkg "github.com/kristianrpo/document-management-microservice/internal/infrastructure/config"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/messaging"
	"github.com/kristianrpo/document-management-microservice/internal/infrastructure/metrics"

	infrapkg "github.com/kristianrpo/document-management-microservice/internal/infrastructure/repository"
)

// @title Document Management Microservice API
// @version 1.0
// @description Microservice for managing document uploads, storage, metadata, and authentication workflows
// @description
// @description Features:
// @description - Upload documents to S3 with automatic storage in DynamoDB
// @description - List, retrieve, and delete documents (individual or bulk)
// @description - Automatic file deduplication based on SHA256 hash
// @description - Support for multiple file types with MIME type detection
// @description - Document authentication workflow via RabbitMQ events
// @description - Transfer documents between operators (generating temporary access links for documents)
// @description - Event-driven architecture for user transfers and authentication results
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
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer <token>" (include the word Bearer followed by a space)
//
// @tag.name documents
// @tag.description Document upload, retrieval, deletion, transfer, and authentication operations
//
// @tag.name health
// @tag.description Health check endpoints
//
//nolint:gocyclo // Main function complexity is acceptable for initialization logic
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using system environment variables")
	}

	config := cfgpkg.Load()

	// Initialize Prometheus metrics
	metricsCollector := metrics.NewPrometheusMetrics("documents_service")
	log.Println("Prometheus metrics initialized")

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

	// Initialize shared RabbitMQ client
	var rabbitMQClient *messaging.RabbitMQClient
	var messagePublisher interfaces.MessagePublisher
	var messageConsumer interfaces.MessageConsumer

	if config.RabbitMQ.URL != "" {
		var err error
		rabbitMQClient, err = messaging.NewRabbitMQClient(config.RabbitMQ)
		if err != nil {
			log.Printf("warning: failed to initialize RabbitMQ client: %v", err)
			log.Println("continuing without RabbitMQ...")
		} else {
			defer func() {
				if err := rabbitMQClient.Close(); err != nil {
					log.Printf("Error closing RabbitMQ client: %v", err)
				}
			}()

			// Initialize publisher with shared client
			rabbitPublisher, err := messaging.NewRabbitMQPublisher(rabbitMQClient)
			if err != nil {
				log.Printf("warning: failed to initialize RabbitMQ publisher: %v", err)
			} else {
				messagePublisher = rabbitPublisher
				log.Println("RabbitMQ publisher initialized")
			}

			// Initialize consumer with shared client
			rabbitConsumer, err := messaging.NewRabbitMQConsumer(rabbitMQClient)
			if err != nil {
				log.Printf("warning: failed to initialize RabbitMQ consumer: %v", err)
			} else {
				messageConsumer = rabbitConsumer
				log.Println("RabbitMQ consumer initialized")
			}
		}
	} else {
		log.Println("RabbitMQ URL not configured, skipping RabbitMQ initialization")
	}

	documentService := usecases.NewDocumentService(
		documentRepository,
		objectStorage,
		fileHasher,
		mimeDetector,
	)
	documentListService := usecases.NewDocumentListService(documentRepository)
	documentGetService := usecases.NewDocumentGetService(documentRepository, objectStorage)
	documentDeleteService := usecases.NewDocumentDeleteService(documentRepository, objectStorage)
	documentDeleteAllService := usecases.NewDocumentDeleteAllService(documentRepository, objectStorage)
	documentTransferService := usecases.NewDocumentTransferService(documentRepository, objectStorage, 15*time.Minute)

	var documentRequestAuthService usecases.DocumentRequestAuthenticationService
	if messagePublisher != nil {
		documentRequestAuthService = usecases.NewDocumentRequestAuthenticationService(
			documentRepository,
			objectStorage,
			messagePublisher,
			config.RabbitMQ.AuthenticationRequestQueue,
			24*time.Hour,
		)
	}

	errorMapper := errors.NewErrorMapper()
	errorHandler := errors.NewErrorHandler(errorMapper)

	uploadHandler := handlers.NewDocumentUploadHandler(documentService, errorHandler, metricsCollector)
	listHandler := handlers.NewDocumentListHandler(documentListService, errorHandler, metricsCollector)

	getHandler := handlers.NewDocumentGetHandler(documentGetService, errorHandler, metricsCollector)
	deleteHandler := handlers.NewDocumentDeleteHandler(documentDeleteService, documentGetService, errorHandler, metricsCollector)
	deleteAllHandler := handlers.NewDocumentDeleteAllHandler(documentDeleteAllService, errorHandler, metricsCollector)
	transferHandler := handlers.NewDocumentTransferHandler(documentTransferService, errorHandler, metricsCollector)

	var requestAuthHandler *handlers.DocumentRequestAuthenticationHandler
	if documentRequestAuthService != nil {
		requestAuthHandler = handlers.NewDocumentRequestAuthenticationHandler(documentRequestAuthService, documentGetService, errorHandler, metricsCollector)
	}

	healthHandler := handlers.NewHealthHandler()

	var jwtMiddleware *middleware.JWTAuthMiddleware
	if config.JWTSecret != "" {
		jwtMiddleware = middleware.NewJWTAuthMiddleware(config.JWTSecret)
		log.Println("JWT middleware initialized")
	} else {
		log.Println("JWT secret not configured; authentication middleware disabled")
	}

	routerConfig := &httpadapter.RouterConfig{
		UploadHandler:      uploadHandler,
		ListHandler:        listHandler,
		GetHandler:         getHandler,
		DeleteHandler:      deleteHandler,
		DeleteAllHandler:   deleteAllHandler,
		TransferHandler:    transferHandler,
		RequestAuthHandler: requestAuthHandler,
		HealthHandler:      healthHandler,
		MetricsCollector:   metricsCollector,
		JWTMiddleware:      jwtMiddleware,
	}

	router := httpadapter.NewRouter(routerConfig)

	// Start consuming messages if messageConsumer is initialized
	ctx := context.Background()
	if messageConsumer != nil {
		// Set up event handlers
		userTransferHandler := events.NewUserTransferHandler(documentDeleteAllService)
		authenticationHandler := events.NewDocumentAuthenticationHandler(documentRepository)
		downloadHandler := events.NewDocumentDownloadHandler(documentService.(interfaces.DocumentUploader), messagePublisher, "documents.ready")

		// Subscribe to user transfer events
		if err := messageConsumer.SubscribeToQueue(ctx, config.RabbitMQ.ConsumerQueue, userTransferHandler.HandleUserTransferred); err != nil {
			log.Printf("warning: failed to subscribe to user transfer queue: %v", err)
		} else {
			log.Printf("listening for user transfer events on queue: %s", config.RabbitMQ.ConsumerQueue)
		}

		// Subscribe to authentication result events
		if err := messageConsumer.SubscribeToQueue(ctx, config.RabbitMQ.AuthenticationResultQueue, authenticationHandler.HandleAuthenticationCompleted); err != nil {
			log.Printf("warning: failed to subscribe to authentication result queue: %v", err)
		} else {
			log.Printf("listening for authentication result events on queue: %s", config.RabbitMQ.AuthenticationResultQueue)
		}

		// Subscribe to document download requested events
		if err := messageConsumer.SubscribeToQueue(ctx, "documents.download.requested", downloadHandler.HandleDownloadRequested); err != nil {
			log.Printf("warning: failed to subscribe to document download requested queue: %v", err)
		} else {
			log.Printf("listening for document download requested events on queue: documents.download.requested")
		}
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
