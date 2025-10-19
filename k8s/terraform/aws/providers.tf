# Locals para reutilizar valores
locals {
  effective_cluster_name = var.cluster_name
  cluster_endpoint       = var.cluster_endpoint
  cluster_ca_data        = var.cluster_ca_certificate
}

# Provider Kubernetes (exec es BLOQUE)
provider "kubernetes" {
  alias                  = "eks"
  host                   = local.cluster_endpoint
  cluster_ca_certificate = base64decode(local.cluster_ca_data)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args        = [
      "eks", "get-token",
      "--cluster-name", local.effective_cluster_name,
      "--region", var.aws_region
    ]
  }
}

# Provider Helm (kubernetes = { ... } es un MAPA; exec también es MAPA)
provider "helm" {
  alias = "eks"

  kubernetes = {
    host                   = local.cluster_endpoint
    cluster_ca_certificate = base64decode(local.cluster_ca_data)
    exec = {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args        = [
        "eks", "get-token",
        "--cluster-name", local.effective_cluster_name,
        "--region", var.aws_region
      ]
    }
  }
}
