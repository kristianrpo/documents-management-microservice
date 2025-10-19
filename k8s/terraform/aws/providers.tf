variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# Si prefieres no pasar variables a mano, lee los outputs de la fase infra:
data "terraform_remote_state" "infra" {
  backend = "local"
  config = {
    path = "${path.module}/../../../infra/terraform/aws/terraform.tfstate"
  }
}

# Variables que por ahora usas en main.tf (puedes dejar ambas opciones)
variable "cluster_name" {
  type = string
  # Si usas remote_state, puedes no pasarla y tomarla de infra:
  default = null
}

variable "aws_lb_controller_role_arn" { type = string }
variable "eso_irsa_role_arn"          { type = string }

locals {
  effective_cluster_name = coalesce(var.cluster_name, try(data.terraform_remote_state.infra.outputs.cluster_name, null))
  cluster_endpoint       = try(data.terraform_remote_state.infra.outputs.cluster_endpoint, null)
  cluster_ca_data_b64    = try(data.terraform_remote_state.infra.outputs.cluster_ca_certificate, null)
}

provider "aws" {
  region = var.aws_region
}

provider "kubernetes" {
  alias                  = "eks"
  host                   = local.cluster_endpoint
  cluster_ca_certificate = base64decode(local.cluster_ca_data_b64)

  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = [
      "eks", "get-token",
      "--cluster-name", local.effective_cluster_name,
      "--region", var.aws_region
    ]
  }
}

provider "helm" {
  alias = "eks"

  # En providers nuevos, "kubernetes" es un OBJETO anidado, no un bloque
  kubernetes = {
    host                   = local.cluster_endpoint
    cluster_ca_certificate = base64decode(local.cluster_ca_data_b64)
    load_config_file       = false
    exec = {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args = [
        "eks", "get-token",
        "--cluster-name", local.effective_cluster_name,
        "--region", var.aws_region
      ]
    }
  }
}
