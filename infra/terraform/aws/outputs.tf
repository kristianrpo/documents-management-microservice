output "kubeconfig" {
  value = {
    name       = module.eks.cluster_name
    endpoint   = module.eks.cluster_endpoint
    ca_data    = module.eks.cluster_certificate_authority_data
    oidc_arn   = module.eks.oidc_provider_arn
    irsa_role  = module.irsa.iam_role_arn
    s3_bucket  = aws_s3_bucket.documents.bucket
    ddb_table  = aws_dynamodb_table.documents.name
  }
  sensitive = true
}
