locals { name = "${var.project}-${var.environment}" }

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = local.name
  cidr = "10.42.0.0/16"

  azs             = ["${var.aws_region}a", "${var.aws_region}b", "${var.aws_region}c"]
  private_subnets = ["10.42.1.0/24", "10.42.2.0/24", "10.42.3.0/24"]
  public_subnets  = ["10.42.101.0/24", "10.42.102.0/24", "10.42.103.0/24"]

  enable_nat_gateway = true
  single_nat_gateway = true

  public_subnet_tags  = { "kubernetes.io/role/elb"          = 1 }
  private_subnet_tags = { "kubernetes.io/role/internal-elb" = 1 }
}

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.8"

  cluster_name                    = local.name
  cluster_version                 = "1.30"
  cluster_endpoint_public_access  = true
  vpc_id                          = module.vpc.vpc_id
  subnet_ids                      = module.vpc.private_subnets
  enable_irsa                     = true

  eks_managed_node_groups = {
    default = {
      min_size       = 2
      max_size       = 4
      desired_size   = 2
      instance_types = ["t3.medium"]
      capacity_type  = "ON_DEMAND"
    }
  }
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
  name         = "Documents"
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


resource "aws_mq_broker" "rabbitmq" {
  broker_name         = "${local.name}-rabbitmq"
  engine_type         = "RabbitMQ"
  engine_version      = "3.13.1"
  host_instance_type  = "mq.t3.micro"
  publicly_accessible = false
  deployment_mode     = "SINGLE_INSTANCE"

  user {
    username = "appuser"
    password = random_password.rabbitmq_password.result
  }

  subnet_ids      = [module.vpc.private_subnets[0]]
  security_groups = [module.vpc.default_security_group_id]

  logs { general = true }
}

resource "random_password" "rabbitmq_password" {
  length           = 20
  special          = true
  override_special = "!@#$%^&*()-_+." # sin [, ], :, =
}

resource "aws_secretsmanager_secret" "app" {
  name        = "${local.name}/application"
  description = "Documents service application configuration"
}

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
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["documents:documents-sa"]
    }
  }
  role_policy_arns = { documents = aws_iam_policy.documents.arn }
}

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

module "irsa_external_secrets" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.39"

  role_name = "${local.name}-eso-irsa"
  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["external-secrets:external-secrets"]
    }
  }
  role_policy_arns = { external_secrets = aws_iam_policy.external_secrets.arn }
}

resource "aws_iam_policy" "aws_load_balancer_controller" {
  name_prefix = "${local.name}-aws-load-balancer-controller-"
  description = "IAM policy for AWS Load Balancer Controller"

  policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
      { Effect="Allow", Action=["iam:CreateServiceLinkedRole"], Resource="*", Condition={ StringEquals={ "iam:AWSServiceName"="elasticloadbalancing.amazonaws.com" } } },
      { Effect="Allow", Action=[
          "ec2:DescribeAccountAttributes","ec2:DescribeAddresses","ec2:DescribeAvailabilityZones",
          "ec2:DescribeInternetGateways","ec2:DescribeVpcs","ec2:DescribeVpcPeeringConnections",
          "ec2:DescribeSubnets","ec2:DescribeSecurityGroups","ec2:DescribeInstances","ec2:DescribeNetworkInterfaces",
          "ec2:DescribeTags","ec2:GetCoipPoolUsage","ec2:DescribeCoipPools",
          "elasticloadbalancing:DescribeLoadBalancers","elasticloadbalancing:DescribeLoadBalancerAttributes",
          "elasticloadbalancing:DescribeListeners","elasticloadbalancing:DescribeListenerCertificates",
          "elasticloadbalancing:DescribeSSLPolicies","elasticloadbalancing:DescribeRules",
          "elasticloadbalancing:DescribeTargetGroups","elasticloadbalancing:DescribeTargetGroupAttributes",
          "elasticloadbalancing:DescribeTargetHealth","elasticloadbalancing:DescribeTags"
        ], Resource="*" },
      { Effect="Allow", Action=[
          "cognito-idp:DescribeUserPoolClient","acm:ListCertificates","acm:DescribeCertificate",
          "iam:ListServerCertificates","iam:GetServerCertificate",
          "waf-regional:GetWebACL","waf-regional:GetWebACLForResource","waf-regional:AssociateWebACL","waf-regional:DisassociateWebACL",
          "wafv2:GetWebACL","wafv2:GetWebACLForResource","wafv2:AssociateWebACL","wafv2:DisassociateWebACL",
          "shield:GetSubscriptionState","shield:DescribeProtection","shield:CreateProtection","shield:DeleteProtection"
        ], Resource="*" },
      { Effect="Allow", Action=["ec2:AuthorizeSecurityGroupIngress","ec2:RevokeSecurityGroupIngress"], Resource="*" },
      { Effect="Allow", Action=["ec2:CreateSecurityGroup"], Resource="*" },
      { Effect="Allow", Action=["ec2:CreateTags"], Resource="arn:aws:ec2:*:*:security-group/*",
        Condition={ StringEquals={ "ec2:CreateAction"="CreateSecurityGroup" }, Null={ "aws:RequestTag/elbv2.k8s.aws/cluster"="false" } } },
      { Effect="Allow", Action=["ec2:CreateTags","ec2:DeleteTags"], Resource="arn:aws:ec2:*:*:security-group/*",
        Condition={ Null={ "aws:RequestTag/elbv2.k8s.aws/cluster"="true","aws:ResourceTag/elbv2.k8s.aws/cluster"="false" } } },
      { Effect="Allow", Action=["ec2:AuthorizeSecurityGroupIngress","ec2:RevokeSecurityGroupIngress","ec2:DeleteSecurityGroup"], Resource="*",
        Condition={ Null={ "aws:ResourceTag/elbv2.k8s.aws/cluster"="false" } } },
      { Effect="Allow", Action=["elasticloadbalancing:CreateLoadBalancer","elasticloadbalancing:CreateTargetGroup"], Resource="*",
        Condition={ Null={ "aws:RequestTag/elbv2.k8s.aws/cluster"="false" } } },
      { Effect="Allow", Action=["elasticloadbalancing:CreateListener","elasticloadbalancing:DeleteListener","elasticloadbalancing:CreateRule","elasticloadbalancing:DeleteRule"], Resource="*" },
      { Effect="Allow", Action=["elasticloadbalancing:AddTags","elasticloadbalancing:RemoveTags"],
        Resource=[
          "arn:aws:elasticloadbalancing:*:*:targetgroup/*/*",
          "arn:aws:elasticloadbalancing:*:*:loadbalancer/net/*/*",
          "arn:aws:elasticloadbalancing:*:*:loadbalancer/app/*/*"
        ],
        Condition={ Null={ "aws:RequestTag/elbv2.k8s.aws/cluster"="true","aws:ResourceTag/elbv2.k8s.aws/cluster"="false" } } },
      { Effect="Allow", Action=["elasticloadbalancing:AddTags","elasticloadbalancing:RemoveTags"],
        Resource=[
          "arn:aws:elasticloadbalancing:*:*:listener/net/*/*/*",
          "arn:aws:elasticloadbalancing:*:*:listener/app/*/*/*",
          "arn:aws:elasticloadbalancing:*:*:listener-rule/net/*/*/*",
          "arn:aws:elasticloadbalancing:*:*:listener-rule/app/*/*/*"
        ] },
      { Effect="Allow", Action=[
          "elasticloadbalancing:ModifyLoadBalancerAttributes","elasticloadbalancing:SetIpAddressType","elasticloadbalancing:SetSecurityGroups",
          "elasticloadbalancing:SetSubnets","elasticloadbalancing:DeleteLoadBalancer","elasticloadbalancing:ModifyTargetGroup",
          "elasticloadbalancing:ModifyTargetGroupAttributes","elasticloadbalancing:DeleteTargetGroup"
        ],
        Resource="*", Condition={ Null={ "aws:ResourceTag/elbv2.k8s.aws/cluster"="false" } } },
      { Effect="Allow", Action=["elasticloadbalancing:AddTags"],
        Resource=[
          "arn:aws:elasticloadbalancing:*:*:targetgroup/*/*",
          "arn:aws:elasticloadbalancing:*:*:loadbalancer/net/*/*",
          "arn:aws:elasticloadbalancing:*:*:loadbalancer/app/*/*"
        ],
        Condition={ StringEquals={ "elasticloadbalancing:CreateAction"=[ "CreateTargetGroup","CreateLoadBalancer" ] },
                    Null={ "aws:RequestTag/elbv2.k8s.aws/cluster"="false" } } },
      { Effect="Allow", Action=["elasticloadbalancing:RegisterTargets","elasticloadbalancing:DeregisterTargets"], Resource="arn:aws:elasticloadbalancing:*:*:targetgroup/*/*" },
      { Effect="Allow", Action=["elasticloadbalancing:SetWebAcl","elasticloadbalancing:ModifyListener","elasticloadbalancing:AddListenerCertificates","elasticloadbalancing:RemoveListenerCertificates","elasticloadbalancing:ModifyRule"], Resource="*" }
    ]
  })
}

module "irsa_aws_load_balancer_controller" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.39"
  role_name = "${local.name}-aws-lb-controller"
  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["kube-system:aws-load-balancer-controller"]
    }
  }
  role_policy_arns = { aws_load_balancer_controller = aws_iam_policy.aws_load_balancer_controller.arn }
}

# Otros outputs útiles
output "s3_bucket" { value = aws_s3_bucket.documents.bucket }
output "dynamodb_table" { value = aws_dynamodb_table.documents.name }
output "rabbitmq_amqp_url" {
  value     = "amqps://appuser:${random_password.rabbitmq_password.result}@${aws_mq_broker.rabbitmq.instances[0].endpoints[0]}/"
  sensitive = true
}
output "irsa_role_arn"               { value = module.irsa.iam_role_arn }
output "secretsmanager_secret_name"  { value = aws_secretsmanager_secret.app.name }
output "secretsmanager_secret_arn"   { value = aws_secretsmanager_secret.app.arn }
output "eso_irsa_role_arn"           { value = module.irsa_external_secrets.iam_role_arn }
output "aws_lb_controller_role_arn"  { value = module.irsa_aws_load_balancer_controller.iam_role_arn }
