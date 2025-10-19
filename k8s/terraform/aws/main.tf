# -------- AWS Load Balancer Controller: SA + chart --------
resource "kubernetes_service_account" "aws_load_balancer_controller" {
  provider = kubernetes.eks
  metadata {
    name      = "aws-load-balancer-controller"
    namespace = "kube-system"
    annotations = {
      "eks.amazonaws.com/role-arn" = var.aws_lb_controller_role_arn
    }
  }
}

resource "helm_release" "aws_load_balancer_controller" {
  provider   = helm.eks
  name       = "aws-load-balancer-controller"
  repository = "https://aws.github.io/eks-charts"
  chart      = "aws-load-balancer-controller"
  namespace  = "kube-system"
  version    = "1.7.1"
  wait       = true
  atomic     = true
  timeout    = 900

  # NUEVA SINTAXIS: listas de objetos
  set = [
    { name = "clusterName",          value = coalesce(var.cluster_name, kubernetes_service_account.aws_load_balancer_controller.metadata[0].namespace) != "" ? local.effective_cluster_name : local.effective_cluster_name },
    { name = "serviceAccount.create", value = "false" },
    { name = "serviceAccount.name",   value = kubernetes_service_account.aws_load_balancer_controller.metadata[0].name }
  ]

  depends_on = [kubernetes_service_account.aws_load_balancer_controller]
}

# -------- External Secrets Operator --------
resource "helm_release" "external_secrets" {
  provider         = helm.eks
  name             = "external-secrets"
  repository       = "https://charts.external-secrets.io"
  chart            = "external-secrets"
  namespace        = "external-secrets"
  create_namespace = true
  version          = "0.9.11"
  wait             = true
  atomic           = true
  timeout          = 900
}

resource "kubernetes_service_account" "external_secrets" {
  provider = kubernetes.eks
  metadata {
    name      = "external-secrets"
    namespace = "external-secrets"
    annotations = {
      "eks.amazonaws.com/role-arn" = var.eso_irsa_role_arn
    }
  }
  depends_on = [helm_release.external_secrets]
}

# -------- Grafana Dashboard (ConfigMap) --------
resource "kubernetes_config_map" "grafana_dashboard" {
  provider = kubernetes.eks
  metadata {
    name      = "documents-service-dashboard"
    namespace = "monitoring"
    labels    = { grafana_dashboard = "1" }
  }

  # Corrijo ruta relativa desde k8s/terraform/aws
  data = {
    "documents-service-dashboard.json" = file("${path.module}/../../../grafana/provisioning/dashboards/documents-service-dashboard.json")
  }
}

# -------- kube-prometheus-stack (Prometheus + Grafana) --------
resource "helm_release" "kube_prometheus_stack" {
  provider         = helm.eks
  name             = "kube-prometheus-stack"
  repository       = "https://prometheus-community.github.io/helm-charts"
  chart            = "kube-prometheus-stack"
  namespace        = "monitoring"
  create_namespace = true
  version          = "56.6.2"
  wait             = true
  atomic           = true
  timeout          = 1200

  # NUEVA SINTAXIS: set / set_sensitive como listas
  set = [
    { name = "prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues", value = "false" },
    { name = "grafana.service.type",                                             value = "LoadBalancer" },
    { name = "grafana.sidecar.dashboards.enabled",                               value = "true" },
    { name = "grafana.sidecar.dashboards.label",                                 value = "grafana_dashboard" }
  ]

  set_sensitive = [
    { name = "grafana.adminPassword", value = "admin" }
  ]

  depends_on = [kubernetes_config_map.grafana_dashboard]
}

# -------- PrometheusRule (CRD) --------
resource "kubernetes_manifest" "prometheus_rules" {
  provider = kubernetes.eks
  manifest = {
    apiVersion = "monitoring.coreos.com/v1"
    kind       = "PrometheusRule"
    metadata = {
      name      = "documents-service-alerts"
      namespace = "monitoring"
      labels    = { prometheus = "kube-prometheus", role = "alert-rules" }
    }
    # Corrijo ruta relativa
    spec = yamldecode(file("${path.module}/../../../prometheus/alerts.yml"))
  }

  # Asegura que la CRD exista antes
  depends_on = [helm_release.kube_prometheus_stack]
}

# -------- Outputs útiles --------
output "kube_prom_stack_status" {
  value = helm_release.kube_prometheus_stack.status
}
