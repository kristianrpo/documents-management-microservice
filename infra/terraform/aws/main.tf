# ============================================================================
# Data source: Consume shared infrastructure from remote state
# ============================================================================
data "terraform_remote_state" "shared" {
  backend = "s3"
  config = {
    bucket = var.tf_backend_bucket
    key    = var.shared_state_key
    region = var.aws_region
  }
}

locals {
  name = "${var.project}-${var.environment}"
  
  # Recursos compartidos desde el remote state
  cluster_name       = data.terraform_remote_state.shared.outputs.cluster_name
  # cluster_oidc_arn   = data.terraform_remote_state.shared.outputs.oidc_provider_arn  # Temporalmente comentado
  rabbitmq_url       = data.terraform_remote_state.shared.outputs.rabbitmq_amqp_url
  processed_messages_table_name = data.terraform_remote_state.shared.outputs.rabbitmq_processed_messages_table_name
  processed_messages_table_arn  = data.terraform_remote_state.shared.outputs.rabbitmq_processed_messages_table_arn
  rabbitmq_consumer_dynamodb_policy_arn = data.terraform_remote_state.shared.outputs.rabbitmq_consumer_dynamodb_policy_arn
  
  # API Gateway outputs from shared infra
  api_gateway_id     = data.terraform_remote_state.shared.outputs.api_gateway_id
  api_gateway_arn    = data.terraform_remote_state.shared.outputs.api_gateway_arn
  vpc_link_id        = data.terraform_remote_state.shared.outputs.api_gateway_vpc_link_id
  api_gateway_stage  = data.terraform_remote_state.shared.outputs.api_gateway_invoke_url
}

# ============================================================================
# Microservice-specific resources
# ============================================================================

resource "random_id" "suffix" {
  byte_length = 2
}

resource "aws_s3_bucket" "documents" {
  bucket_prefix = "${local.name}-"
  force_destroy = true
}

resource "aws_s3_bucket_public_access_block" "this" {
  bucket = aws_s3_bucket.documents.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_dynamodb_table" "documents" {
  name         = "${local.name}-documents-${random_id.suffix.hex}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "DocumentID"
  range_key    = "OwnerID"

  attribute {
    name = "DocumentID"
    type = "S"
  }
  attribute {
    name = "OwnerID"
    type = "N"
  }
  attribute {
    name = "HashSHA256"
    type = "S"
  }

  global_secondary_index {
    name            = "OwnerIDIndex"
    hash_key        = "OwnerID"
    projection_type = "ALL"
  }

  global_secondary_index {
    name            = "HashOwnerIndex"
    hash_key        = "HashSHA256"
    range_key       = "OwnerID"
    projection_type = "ALL"
  }
}

# ============================================================================
# Secret Manager for application config
# ============================================================================
resource "aws_secretsmanager_secret" "app" {
  name        = "${local.name}/application-${random_id.suffix.hex}"
  description = "Documents service application configuration"
}

# ============================================================================
# IAM Role for Documents Service (IRSA)
# ============================================================================
data "aws_iam_policy_document" "documents_policy" {
  statement {
    actions   = ["s3:PutObject","s3:GetObject","s3:DeleteObject","s3:ListBucket"]
    resources = [aws_s3_bucket.documents.arn, "${aws_s3_bucket.documents.arn}/*"]
  }
  statement {
    actions   = ["dynamodb:PutItem","dynamodb:GetItem","dynamodb:DeleteItem","dynamodb:Query","dynamodb:BatchWriteItem","dynamodb:UpdateItem"]
    resources = [aws_dynamodb_table.documents.arn, "${aws_dynamodb_table.documents.arn}/index/*"]
  }
  statement {
    actions   = ["dynamodb:PutItem","dynamodb:GetItem","dynamodb:Query"]
    resources = [local.processed_messages_table_arn]
  }
  statement {
    actions   = ["secretsmanager:GetSecretValue"]
    resources = [aws_secretsmanager_secret.app.arn]
  }
}

resource "aws_iam_policy" "documents" {
  name_prefix = "${local.name}-policy-"
  policy = data.aws_iam_policy_document.documents_policy.json
}

module "irsa" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.39"

  role_name = "${local.name}-documents-irsa"
  oidc_providers = {
    main = {
      provider_arn               = data.terraform_remote_state.shared.outputs.oidc_provider_arn
      namespace_service_accounts = ["documents:documents-sa"]
    }
  }
  role_policy_arns = { 
    documents = aws_iam_policy.documents.arn
    rabbitmq_consumer = local.rabbitmq_consumer_dynamodb_policy_arn
  }
}

# ============================================================================
# IAM Policy for External Secrets to access this service's secret
# (attached to shared ESO role via inline policy or additional policy)
# ============================================================================
data "aws_iam_policy_document" "external_secrets" {
  statement {
    actions   = ["secretsmanager:GetSecretValue"]
    resources = [aws_secretsmanager_secret.app.arn]
  }
}

resource "aws_iam_policy" "external_secrets" {
  name_prefix = "${local.name}-external-secrets-policy-"
  policy = data.aws_iam_policy_document.external_secrets.json
}

# Attach this policy to the shared ESO role from shared infra
# This allows External Secrets Operator to read THIS microservice's secret
resource "aws_iam_role_policy_attachment" "eso_documents_secret" {
  role       = data.terraform_remote_state.shared.outputs.eso_irsa_role_name
  policy_arn = aws_iam_policy.external_secrets.arn
}

# ============================================================================
# API GATEWAY INTEGRATION
# ============================================================================
# This integrates this microservice's ALB with the shared API Gateway

# Data source: Find the ALB created by AWS Load Balancer Controller
# The ALB is tagged by the Kubernetes ingress annotations
# Note: This ALB is created dynamically by the AWS Load Balancer Controller
# and will be available after the ingress is deployed in Kubernetes
data "aws_lb" "documents_alb" {
  tags = {
    Service     = "documents"
    Environment = "prod"
  }
}

# Data source: Get VPC Link to find its security group
data "aws_apigatewayv2_vpc_link" "api_gateway_vpc_link" {
  vpc_link_id = local.vpc_link_id
}

# Security Group Rule: Allow VPC Link to access ALB
# This is required for API Gateway to reach the ALB through VPC Link
resource "aws_security_group_rule" "vpc_link_to_alb" {
  type                     = "ingress"
  from_port                = 80
  to_port                  = 80
  protocol                 = "tcp"
  source_security_group_id = tolist(data.aws_apigatewayv2_vpc_link.api_gateway_vpc_link.security_group_ids)[0]
  security_group_id        = tolist(data.aws_lb.documents_alb.security_groups)[0]
  description              = "Allow API Gateway VPC Link to access ALB"
}

# Data source: Get the HTTP listener (port 80) of the ALB
# API Gateway needs the listener ARN, not the DNS name
data "aws_lb_listener" "documents_alb_http" {
  load_balancer_arn = data.aws_lb.documents_alb.arn
  port              = 80
}

# API Gateway Integration: Connects API Gateway to the ALB via VPC Link
# Uses the ALB listener ARN - the listener handles all routing
resource "aws_apigatewayv2_integration" "documents" {
  api_id           = local.api_gateway_id
  integration_type = "HTTP_PROXY"
  
  connection_type        = "VPC_LINK"
  connection_id          = local.vpc_link_id
  integration_method     = "ANY"
  integration_uri        = data.aws_lb_listener.documents_alb_http.arn
  payload_format_version = "1.0"
}

# API Gateway Route: /api/docs/*
# Routes to the documents microservice via ALB
resource "aws_apigatewayv2_route" "documents_api" {
  api_id    = local.api_gateway_id
  route_key = "ANY /api/docs/{proxy+}"
  
  target = "integrations/${aws_apigatewayv2_integration.documents.id}"
}

# ============================================================================
# Outputs - Only microservice-specific resources
# ============================================================================
output "s3_bucket"                 { value = aws_s3_bucket.documents.bucket }
output "dynamodb_table"            { value = aws_dynamodb_table.documents.name }
output "rabbitmq_amqp_url"         { 
  value     = local.rabbitmq_url
  sensitive = true
}
output "rabbitmq_processed_messages_table_name" {
  value = local.processed_messages_table_name
}
output "irsa_role_arn"             { value = module.irsa.iam_role_arn }
output "secretsmanager_secret_name"{ value = aws_secretsmanager_secret.app.name }
output "secretsmanager_secret_arn" { value = aws_secretsmanager_secret.app.arn }

# Outputs from shared infrastructure (for convenience)
output "cluster_name"              { value = local.cluster_name }
output "cluster_endpoint"          { value = data.terraform_remote_state.shared.outputs.cluster_endpoint }
output "cluster_ca_certificate"    { value = data.terraform_remote_state.shared.outputs.cluster_ca_certificate }
output "aws_lb_controller_role_arn"{ value = data.terraform_remote_state.shared.outputs.aws_load_balancer_controller_irsa_role_arn }

# API Gateway outputs
output "alb_hostname" {
  description = "ALB hostname for this microservice"
  value       = try(data.aws_lb.documents_alb.dns_name, "Pending ALB creation")
}

output "api_gateway_url" {
  description = "API Gateway URL for this microservice"
  value       = "${local.api_gateway_stage}/api/docs"
}

output "api_gateway_health_check_url" {
  description = "Health check URL via API Gateway"
  value       = "${local.api_gateway_stage}/api/docs/healthz"
}

output "api_gateway_swagger_url" {
  description = "Swagger documentation URL via API Gateway"
  value       = "${local.api_gateway_stage}/api/docs/swagger/"
}
