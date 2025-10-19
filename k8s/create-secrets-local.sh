#!/bin/bash
# create-secrets-local.sh
# Creates a single Kubernetes Secret with ALL environment variables
# Simulates what an external secret manager (AWS Secrets Manager, Vault, etc.) would inject
#
# ⚠️  LOCAL DEVELOPMENT ONLY - uses test credentials
# For production/AWS, use:
#   - External Secrets Operator (syncs from AWS Secrets Manager)
#   - GitHub Actions with secrets injected from GitHub Secrets
#   - IRSA for AWS credentials (no static keys)

set -e

NAMESPACE="documents"

echo "🔐 Creating consolidated secret for LOCAL development in namespace: $NAMESPACE"
echo "   (simulating external secret injection)"
echo ""

# Create namespace if it doesn't exist
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Single secret with ALL environment variables
# In production, this would come from AWS Secrets Manager or similar
kubectl -n $NAMESPACE create secret generic documents-secrets \
  --from-literal=APP_PORT="8080" \
  --from-literal=AWS_REGION="us-east-1" \
  --from-literal=AWS_ACCESS_KEY_ID="admin" \
  --from-literal=AWS_SECRET_ACCESS_KEY="admin123" \
  --from-literal=DYNAMODB_TABLE="Documents" \
  --from-literal=DYNAMODB_ENDPOINT="http://dynamodb-local:8000" \
  --from-literal=S3_BUCKET="documents" \
  --from-literal=S3_ENDPOINT="http://minio:9000" \
  --from-literal=S3_USE_PATH_STYLE="true" \
  --from-literal=S3_PUBLIC_BASE_URL="http://localhost:9000/documents" \
  --from-literal=RABBITMQ_URL="amqp://guest:guest@rabbitmq:5672/" \
  --from-literal=RABBITMQ_CONSUMER_QUEUE="user.transferred" \
  --from-literal=RABBITMQ_AUTH_REQUEST_QUEUE="document.authentication.requested" \
  --from-literal=RABBITMQ_AUTH_RESULT_QUEUE="document.authentication.completed" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "✅ Secret 'documents-secrets' created with all environment variables!"
echo ""
echo "📋 Secret contains:"
echo "   - AWS credentials (admin/admin123 for local MinIO/DynamoDB)"
echo "   - S3 configuration (MinIO endpoint)"
echo "   - DynamoDB configuration (local endpoint)"
echo "   - RabbitMQ URL (guest/guest@rabbitmq:5672)"
echo "   - Queue names and app settings"
echo ""
echo "Next steps:"
echo "  1. Build image:      docker build -t documents-service:local ."
echo "  2. Load to cluster:  kind load docker-image documents-service:local --name <cluster-name>"
echo "  3. Deploy:           kubectl apply -k k8s/"
echo ""
echo "🔍 To view secret (base64 decoded):"
echo "   kubectl -n documents get secret documents-secrets -o yaml"
echo ""
