package repository

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain"
)

type dynamoDBDocumentRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewDynamoDBDocumentRepo(client *dynamodb.Client, tableName string) interfaces.DocumentRepository {
	return &dynamoDBDocumentRepository{
		client:    client,
		tableName: tableName,
	}
}

func (repo *dynamoDBDocumentRepository) Create(document *domain.Document) error {
	if document.ID == "" {
		document.ID = uuid.New().String()
	}
	now := time.Now()
	if document.CreatedAt.IsZero() {
		document.CreatedAt = now
	}
	document.UpdatedAt = now

	item, err := attributevalue.MarshalMap(document)
	if err != nil {
		return err
	}

	_, err = repo.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(repo.tableName),
		Item:      item,
	})

	return err
}
