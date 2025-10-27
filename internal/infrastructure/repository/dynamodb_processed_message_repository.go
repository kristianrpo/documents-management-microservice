package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

// DynamoDBProcessedMessageRepository implements ProcessedMessageRepository using DynamoDB
type DynamoDBProcessedMessageRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBProcessedMessageRepository creates a new DynamoDB-based processed message repository
func NewDynamoDBProcessedMessageRepository(client *dynamodb.Client, tableName string) interfaces.ProcessedMessageRepository {
	return &DynamoDBProcessedMessageRepository{
		client:    client,
		tableName: tableName,
	}
}

// CheckIfProcessed checks if a message has been processed by querying DynamoDB
func (r *DynamoDBProcessedMessageRepository) CheckIfProcessed(ctx context.Context, messageID string) (bool, error) {
	if messageID == "" {
		return false, nil
	}

	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"MessageID": &types.AttributeValueMemberS{Value: messageID},
		},
	})

	if err != nil {
		return false, fmt.Errorf("failed to check if message is processed: %w", err)
	}

	// If no item found, message hasn't been processed
	if result.Item == nil {
		return false, nil
	}

	return true, nil
}

// MarkAsProcessed marks a message as processed in DynamoDB
func (r *DynamoDBProcessedMessageRepository) MarkAsProcessed(ctx context.Context, message *models.ProcessedMessage) error {
	if message == nil {
		return fmt.Errorf("message cannot be nil")
	}

	// Marshal the message to DynamoDB attributes
	item, err := attributevalue.MarshalMap(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Put the item (will overwrite if exists, which is fine for idempotency)
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to mark message as processed: %w", err)
	}

	return nil
}

// EnsureTableExists creates the processed messages table if it doesn't exist
func (r *DynamoDBProcessedMessageRepository) EnsureTableExists(ctx context.Context) error {
	// Check if table already exists
	_, err := r.client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(r.tableName),
	})

	if err == nil {
		// Table exists
		return nil
	}

	// Table doesn't exist, create it
	_, err = r.client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:   aws.String(r.tableName),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("MessageID"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("MessageID"),
				KeyType:       types.KeyTypeHash,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create processed messages table: %w", err)
	}

	// Wait for table to be active
	waiter := dynamodb.NewTableExistsWaiter(r.client)
	return waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(r.tableName),
	}, time.Second*30)
}
