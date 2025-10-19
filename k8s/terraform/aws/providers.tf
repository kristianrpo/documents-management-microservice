# Provider AWS para datasources (token/endpoint/CA)
provider "aws" {
  region = var.aws_region
}

# Datasources EKS (permiten obtener endpoint/CA y token)
data "aws_eks_cluster" "this" {
  count = var.cluster_name == null ? 0 : 1
  name  = var.cluster_name
}

data "aws_eks_cluster_auth" "this" {
  count = var.cluster_name == null ? 0 : 1
  name  = var.cluster_name
}

locals {
  effective_cluster_name     = var.cluster_name
  effective_cluster_endpoint = coalesce(var.cluster_endpoint, try(data.aws_eks_cluster.this[0].endpoint, null))
  effective_cluster_ca_data  = coalesce(var.cluster_ca_certificate, try(data.aws_eks_cluster.this[0].certificate_authority[0].data, null))
}

provider "kubernetes" {
  alias                  = "eks"
  host                   = local.effective_cluster_endpoint
  cluster_ca_certificate = base64decode(local.effective_cluster_ca_data)
  token                  = try(data.aws_eks_cluster_auth.this[0].token, null)
}

provider "helm" {
  alias = "eks"

  kubernetes = {
    host                   = local.effective_cluster_endpoint
    cluster_ca_certificate = base64decode(local.effective_cluster_ca_data)
    token                  = try(data.aws_eks_cluster_auth.this[0].token, null)
  }
}
