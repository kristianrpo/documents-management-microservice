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
		log.Printf("dynamodb Query error on GSI HashEmailIndex: %v", err)
		return nil, nil
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var document domain.Document
	if err := attributevalue.UnmarshalMap(result.Items[0], &document); err != nil {
		return nil, err
	}
	return &document, nil
}

func (repo *dynamoDBDocumentRepository) GetByID(id string) (*domain.Document, error) {
	result, err := repo.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("DocumentID = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: id},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		log.Printf("dynamodb Query error: %v", err)
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var document domain.Document
	if err := attributevalue.UnmarshalMap(result.Items[0], &document); err != nil {
		return nil, err
	}
	return &document, nil
}

func (repo *dynamoDBDocumentRepository) List(ownerEmail string, limit, offset int) ([]*domain.Document, int64, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		IndexName:              aws.String("OwnerEmailIndex"),
		KeyConditionExpression: aws.String("OwnerEmail = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: ownerEmail},
		},
		ScanIndexForward: aws.Bool(false),
	}

	var allDocuments []*domain.Document
	var lastEvaluatedKey map[string]types.AttributeValue

	for {
		if lastEvaluatedKey != nil {
			queryInput.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := repo.client.Query(context.TODO(), queryInput)
		if err != nil {
			log.Printf("dynamodb Query error on GSI OwnerEmailIndex: %v", err)
			return nil, 0, err
		}

		for _, item := range result.Items {
			var doc domain.Document
			if err := attributevalue.UnmarshalMap(item, &doc); err != nil {
				log.Printf("error unmarshaling document: %v", err)
				continue
			}
			allDocuments = append(allDocuments, &doc)
		}

		lastEvaluatedKey = result.LastEvaluatedKey
		if lastEvaluatedKey == nil {
			break
		}

		if len(allDocuments) >= offset+limit {
			break
		}
	}

	totalCount := int64(len(allDocuments))

	start := offset
	if start > len(allDocuments) {
		return []*domain.Document{}, totalCount, nil
	}

	end := start + limit
	if end > len(allDocuments) {
		end = len(allDocuments)
	}

	paginatedDocuments := allDocuments[start:end]
	return paginatedDocuments, totalCount, nil
}

func (repo *dynamoDBDocumentRepository) DeleteByID(id string) (*domain.Document, error) {
	document, err := repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if document == nil {
		return nil, nil
	}

	_, err = repo.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"DocumentID": &types.AttributeValueMemberS{Value: document.ID},
			"OwnerEmail": &types.AttributeValueMemberS{Value: document.OwnerEmail},
		},
	})
	if err != nil {
		log.Printf("dynamodb DeleteItem error: %v", err)
		return nil, err
	}

	return document, nil
}

func (repo *dynamoDBDocumentRepository) DeleteAllByEmail(email string) (int, error) {
	documents, _, err := repo.List(email, 1000, 0)
	if err != nil {
		log.Printf("dynamodb List error during DeleteAllByEmail: %v", err)
		return 0, err
	}

	if len(documents) == 0 {
		return 0, nil
	}

	deletedCount := 0
	for _, doc := range documents {
		_, err := repo.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
			TableName: aws.String(repo.tableName),
			Key: map[string]types.AttributeValue{
				"DocumentID": &types.AttributeValueMemberS{Value: doc.ID},
				"OwnerEmail": &types.AttributeValueMemberS{Value: doc.OwnerEmail},
			},
		})
		if err != nil {
			log.Printf("dynamodb DeleteItem error for document %s: %v", doc.ID, err)
			continue
		}
		deletedCount++
	}

	return deletedCount, nil
}
