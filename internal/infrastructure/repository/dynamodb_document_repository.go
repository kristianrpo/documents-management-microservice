package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/models"
)

const (
	// GSI index names
	hashOwnerIndexName = "HashOwnerIndex"
	ownerIDIndexName   = "OwnerIDIndex"

	// Batch operation limits
	maxBatchDeleteSize = 25 // DynamoDB BatchWriteItem limit
	bulkQueryLimit     = 1000

	// DynamoDB expression attribute names
	ownerIDAttr = ":ownerid"

	// Error messages
	errUnmarshalDocument = "failed to unmarshal document: %w"
)

// dynamoDBDocumentRepository implements the DocumentRepository interface using AWS DynamoDB
type dynamoDBDocumentRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBDocumentRepo creates a new DynamoDB document repository
func NewDynamoDBDocumentRepo(client *dynamodb.Client, tableName string) interfaces.DocumentRepository {
	return &dynamoDBDocumentRepository{
		client:    client,
		tableName: tableName,
	}
}

// Create stores a new document in DynamoDB, generating an ID and timestamps if not present
func (repo *dynamoDBDocumentRepository) Create(ctx context.Context, document *models.Document) error {
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
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	_, err = repo.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(repo.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create document in DynamoDB: %w", err)
	}

	return nil
}

// FindByHashAndOwnerID retrieves a document by its hash and owner ID using the HashOwnerIndex GSI
// This is used for file deduplication
func (repo *dynamoDBDocumentRepository) FindByHashAndOwnerID(ctx context.Context, hashSHA256 string, ownerID int64) (*models.Document, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		IndexName:              aws.String(hashOwnerIndexName),
		KeyConditionExpression: aws.String("HashSHA256 = :hash AND OwnerID = " + ownerIDAttr),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":hash":     &types.AttributeValueMemberS{Value: hashSHA256},
			ownerIDAttr: &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", ownerID)},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query document by hash and owner: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var document models.Document
	if err := attributevalue.UnmarshalMap(result.Items[0], &document); err != nil {
		return nil, fmt.Errorf(errUnmarshalDocument, err)
	}
	return &document, nil
}

// GetByID retrieves a document by its unique identifier
func (repo *dynamoDBDocumentRepository) GetByID(ctx context.Context, id string) (*models.Document, error) {
	result, err := repo.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		KeyConditionExpression: aws.String("DocumentID = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: id},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get document by ID: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var document models.Document
	if err := attributevalue.UnmarshalMap(result.Items[0], &document); err != nil {
		return nil, fmt.Errorf(errUnmarshalDocument, err)
	}
	return &document, nil
}

// List retrieves a paginated list of documents for a specific owner using the OwnerIDIndex GSI
// Returns documents sorted by creation date (most recent first)
func (repo *dynamoDBDocumentRepository) List(ctx context.Context, ownerID int64, limit, offset int) ([]*models.Document, int64, error) {
	totalCount, err := repo.countDocumentsByOwner(ctx, ownerID)
	if err != nil {
		return nil, 0, err
	}

	if offset >= int(totalCount) {
		return []*models.Document{}, totalCount, nil
	}

	documents, err := repo.fetchPaginatedDocuments(ctx, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return documents, totalCount, nil
}

// countDocumentsByOwner returns the total count of documents for a specific owner
func (repo *dynamoDBDocumentRepository) countDocumentsByOwner(ctx context.Context, ownerID int64) (int64, error) {
	countInput := &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		IndexName:              aws.String(ownerIDIndexName),
		KeyConditionExpression: aws.String("OwnerID = " + ownerIDAttr),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			ownerIDAttr: &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", ownerID)},
		},
		Select: types.SelectCount,
	}

	countResult, err := repo.client.Query(ctx, countInput)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}
	return int64(countResult.Count), nil
}

// fetchPaginatedDocuments retrieves paginated documents for an owner
func (repo *dynamoDBDocumentRepository) fetchPaginatedDocuments(ctx context.Context, ownerID int64, limit, offset int) ([]*models.Document, error) {
	queryInput := repo.buildQueryInput(ownerID)
	
	var documents []*models.Document
	pagination := &paginationState{
		limit:  limit,
		offset: offset,
	}

	for !pagination.isDone() {
		result, err := repo.executeQuery(ctx, queryInput)
		if err != nil {
			return nil, err
		}

		documents, err = repo.processQueryResults(result.Items, documents, pagination)
		if err != nil {
			return nil, err
		}

		if result.LastEvaluatedKey == nil || pagination.isDone() {
			break
		}
		queryInput.ExclusiveStartKey = result.LastEvaluatedKey
	}

	return documents, nil
}

// paginationState tracks pagination progress
type paginationState struct {
	limit          int
	offset         int
	itemsSkipped   int
	itemsCollected int
}

func (p *paginationState) isDone() bool {
	return p.itemsCollected >= p.limit
}

func (p *paginationState) shouldSkip() bool {
	if p.itemsSkipped < p.offset {
		p.itemsSkipped++
		return true
	}
	return false
}

func (p *paginationState) collect() {
	p.itemsCollected++
}

// buildQueryInput creates a query input for fetching documents by owner
func (repo *dynamoDBDocumentRepository) buildQueryInput(ownerID int64) *dynamodb.QueryInput {
	return &dynamodb.QueryInput{
		TableName:              aws.String(repo.tableName),
		IndexName:              aws.String(ownerIDIndexName),
		KeyConditionExpression: aws.String("OwnerID = " + ownerIDAttr),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			ownerIDAttr: &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", ownerID)},
		},
		ScanIndexForward: aws.Bool(false),
	}
}

// executeQuery executes a DynamoDB query
func (repo *dynamoDBDocumentRepository) executeQuery(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	result, err := repo.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	return result, nil
}

// processQueryResults processes query items and applies pagination
func (repo *dynamoDBDocumentRepository) processQueryResults(items []map[string]types.AttributeValue, documents []*models.Document, pagination *paginationState) ([]*models.Document, error) {
	for _, item := range items {
		if pagination.shouldSkip() {
			continue
		}

		if pagination.isDone() {
			return documents, nil
		}

		var doc models.Document
		if err := attributevalue.UnmarshalMap(item, &doc); err != nil {
			return nil, fmt.Errorf(errUnmarshalDocument, err)
		}
		documents = append(documents, &doc)
		pagination.collect()
	}
	return documents, nil
}

// DeleteByID removes a document by its ID and returns the deleted document
// Returns nil if the document doesn't exist
func (repo *dynamoDBDocumentRepository) DeleteByID(ctx context.Context, id string) (*models.Document, error) {
	document, err := repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if document == nil {
		return nil, nil
	}

	_, err = repo.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"DocumentID": &types.AttributeValueMemberS{Value: document.ID},
			"OwnerID":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", document.OwnerID)},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	return document, nil
}

// DeleteAllByOwnerID removes all documents owned by a specific user
// Uses batch operations for efficiency (max 25 items per batch)
func (repo *dynamoDBDocumentRepository) DeleteAllByOwnerID(ctx context.Context, ownerID int64) (int, error) {
	documents, _, err := repo.List(ctx, ownerID, bulkQueryLimit, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to list documents for deletion: %w", err)
	}

	if len(documents) == 0 {
		return 0, nil
	}

	deletedCount := 0

	for i := 0; i < len(documents); i += maxBatchDeleteSize {
		end := i + maxBatchDeleteSize
		if end > len(documents) {
			end = len(documents)
		}
		batch := documents[i:end]

		var writeRequests []types.WriteRequest
		for _, doc := range batch {
			writeRequests = append(writeRequests, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						"DocumentID": &types.AttributeValueMemberS{Value: doc.ID},
						"OwnerID":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", doc.OwnerID)},
					},
				},
			})
		}

		_, err := repo.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				repo.tableName: writeRequests,
			},
		})
		if err != nil {
			return deletedCount, fmt.Errorf("failed to batch delete documents: %w", err)
		}

		deletedCount += len(batch)
	}

	return deletedCount, nil
}

// UpdateAuthenticationStatus updates the authentication status of a document and its updated timestamp
func (repo *dynamoDBDocumentRepository) UpdateAuthenticationStatus(ctx context.Context, documentID string, status models.AuthenticationStatus) error {
	now := time.Now()

	document, err := repo.GetByID(ctx, documentID)
	if err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}

	if document == nil {
		return fmt.Errorf("document not found")
	}

	_, err = repo.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(repo.tableName),
		Key: map[string]types.AttributeValue{
			"DocumentID": &types.AttributeValueMemberS{Value: documentID},
			"OwnerID":    &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", document.OwnerID)},
		},
		UpdateExpression: aws.String("SET AuthenticationStatus = :status, UpdatedAt = :updated"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":  &types.AttributeValueMemberS{Value: string(status)},
			":updated": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339Nano)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update authentication status: %w", err)
	}

	return nil
}
