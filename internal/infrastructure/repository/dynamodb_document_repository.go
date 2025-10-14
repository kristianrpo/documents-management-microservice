package repository

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

	if err != nil {
		log.Printf("dynamodb PutItem error: %v", err)
	}
	return err
}

func (repo *dynamoDBDocumentRepository) FindByHashAndEmail(hashSHA256, ownerEmail string) (*domain.Document, error) {
	result, err := repo.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		IndexName:              aws.String("HashEmailIndex"),
		KeyConditionExpression: aws.String("HashSHA256 = :hash AND OwnerEmail = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hash":  &types.AttributeValueMemberS{Value: hashSHA256},
			":email": &types.AttributeValueMemberS{Value: ownerEmail},
		},
		Limit: aws.Int32(1),
	})

	if err != nil {
		// Log query error; if index doesn't exist locally, treat as no result
		log.Printf("dynamodb Query error on GSI HashEmailIndex: %v", err)
		return nil, nil
	}

	if len(result.Items) == 0 {
		return nil, nil // No document found
	}

	var document domain.Document
	err = attributevalue.UnmarshalMap(result.Items[0], &document)
	if err != nil {
		return nil, err
	}

	return &document, nil
}
