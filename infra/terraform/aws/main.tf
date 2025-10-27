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

# Attach this policy to the shared ESO role
# resource "aws_iam_role_policy_attachment" "eso_documents_secret" {
#   role       = data.terraform_remote_state.shared.outputs.external_secrets_role_name
#   policy_arn = aws_iam_policy.external_secrets.arn
# }

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
output "eso_irsa_role_arn"         { value = data.terraform_remote_state.shared.outputs.external_secrets_irsa_role_arn }
output "aws_lb_controller_role_arn"{ value = data.terraform_remote_state.shared.outputs.aws_load_balancer_controller_irsa_role_arn }
