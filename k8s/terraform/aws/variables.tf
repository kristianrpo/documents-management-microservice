variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# EKS cluster info (from shared infrastructure outputs or local)
variable "cluster_name" {
  type    = string
  default = null
}

variable "cluster_endpoint" {
  type    = string
  default = null
}

variable "cluster_ca_certificate" {
  type    = string
  default = null
}
