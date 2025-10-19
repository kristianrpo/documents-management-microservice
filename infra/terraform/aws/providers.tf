variable "aws_region" {
  type    = string
  default = "us-east-1"
}
variable "project" {
  type    = string
  default = "documents"
}
variable "environment" {
  type    = string
  default = "dev"
}

provider "aws" {
  region = var.aws_region
}

# Salidas útiles para la fase 2
output "cluster_name" {
  value = module.eks.cluster_name
}
output "cluster_endpoint" {
  value = module.eks.cluster_endpoint
}
output "cluster_ca_certificate" {
  value = module.eks.cluster_certificate_authority_data
}
