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

variable "tf_backend_bucket" {
  type        = string
  description = "S3 bucket for terraform remote state (same bucket used for both shared and microservice state)"
}

variable "shared_state_key" {
  type        = string
  default     = "shared/terraform.tfstate"
  description = "S3 key where shared infrastructure state is stored"
}

provider "aws" {
  region = var.aws_region
}
