#!/bin/bash

# This script initializes the DynamoDB table for local development.
# Ensure that you have the AWS CLI and DynamoDB Local running before executing this script.

# Table name and region configuration
TABLE_NAME="Documents"
REGION="us-west-2"

# Check if DynamoDB Local is running
if ! nc -z localhost 8000; then
  echo "DynamoDB Local is not running. Please start it before running this script."
  exit 1
fi

# Delete the table if it already exists
echo "Deleting existing table (if any)..."
aws dynamodb delete-table \
  --table-name $TABLE_NAME \
  --endpoint-url http://localhost:8000 \
  --region $REGION

# Wait for the table to be deleted
aws dynamodb wait table-not-exists \
  --table-name $TABLE_NAME \
  --endpoint-url http://localhost:8000 \
  --region $REGION

echo "Creating table $TABLE_NAME..."

# Create the table with both GSIs
aws dynamodb create-table \
  --table-name $TABLE_NAME \
  --attribute-definitions \
    AttributeName=DocumentID,AttributeType=S \
    AttributeName=OwnerEmail,AttributeType=S \
    AttributeName=HashSHA256,AttributeType=S \
  --key-schema \
    AttributeName=DocumentID,KeyType=HASH \
    AttributeName=OwnerEmail,KeyType=RANGE \
  --provisioned-throughput \
    ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --global-secondary-indexes \
    "[
      {
        \"IndexName\": \"OwnerEmailIndex\",
        \"KeySchema\": [
          {\"AttributeName\": \"OwnerEmail\", \"KeyType\": \"HASH\"}
        ],
        \"Projection\": {\"ProjectionType\": \"ALL\"},
        \"ProvisionedThroughput\": {
          \"ReadCapacityUnits\": 5,
          \"WriteCapacityUnits\": 5
        }
      },
      {
        \"IndexName\": \"HashEmailIndex\",
        \"KeySchema\": [
          {\"AttributeName\": \"HashSHA256\", \"KeyType\": \"HASH\"},
          {\"AttributeName\": \"OwnerEmail\", \"KeyType\": \"RANGE\"}
        ],
        \"Projection\": {\"ProjectionType\": \"ALL\"},
        \"ProvisionedThroughput\": {
          \"ReadCapacityUnits\": 5,
          \"WriteCapacityUnits\": 5
        }
      }
    ]" \
  --endpoint-url http://localhost:8000 \
  --region $REGION

# Wait for the table to be created
aws dynamodb wait table-exists \
  --table-name $TABLE_NAME \
  --endpoint-url http://localhost:8000 \
  --region $REGION

echo "Table $TABLE_NAME has been created successfully with GSIs: OwnerEmailIndex and HashEmailIndex."