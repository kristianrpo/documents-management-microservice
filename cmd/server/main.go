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

	httpadapter "github.com/kristianrpo/document-management-microservice/internal/adapters/http"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/errors"
	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/handlers"
	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/application/usecases"
	"github.com/kristianrpo/document-management-microservice/internal/application/util"
	cfgpkg "github.com/kristianrpo/document-management-microservice/internal/infrastructure/config"
	infrapkg "github.com/kristianrpo/document-management-microservice/internal/infrastructure/repository"
	storagepkg "github.com/kristianrpo/document-management-microservice/internal/infrastructure/storage"
)

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

	s3Client, err := storagepkg.NewS3(context.Background(), storagepkg.S3Opts{
		AccessKey:    config.AWSAccessKey,
		SecretKey:    config.AWSSecretKey,
		Region:       config.AWSRegion,
		Endpoint:     config.S3Endpoint,
		Bucket:       config.S3Bucket,
		UsePathStyle: config.S3UsePath,
		PublicBase:   config.S3PublicBase,
	})
	if err != nil {
		log.Fatalf("s3 init: %v", err)
	}

	documentRepository := infrapkg.NewDynamoDBDocumentRepo(dynamoClient, config.DynamoDBTable)
	
	var objectStorage interfaces.ObjectStorage = s3Client
	
	fileHasher := util.NewSHA256Hasher()
	mimeDetector := util.NewExtensionBasedDetector()
	
	documentService := usecases.NewDocumentService(
		documentRepository,
		objectStorage,
		fileHasher,
		mimeDetector,
	)

	errorMapper := errors.NewErrorMapper()
	errorHandler := errors.NewErrorHandler(errorMapper)

	uploadHandler := handlers.NewDocumentUploadHandler(documentService, errorHandler)
	healthHandler := handlers.NewHealthHandler()

	router := httpadapter.NewRouter(uploadHandler, healthHandler)

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
