variable "aws_region" {
  type    = string
  default = "us-east-1"
}

# Estos 3 vienen de la Fase 1 (outputs remotos en tu pipeline)
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

# IRSA de fase 1 (outputs)
variable "aws_lb_controller_role_arn" {
  type    = string
  default = null
}

variable "eso_irsa_role_arn" {
  type    = string
  default = null
}
