# ============================================================================
# Microservice-specific K8s resources
# Note: Prometheus, Grafana, AWS Load Balancer Controller and External Secrets Operator
# are now deployed by the shared infrastructure repository
# ============================================================================

# -------- Grafana Dashboard (ConfigMap) --------
# Deploy to the monitoring namespace created by shared infrastructure
resource "kubernetes_config_map" "grafana_dashboard" {
  provider = kubernetes.eks
  metadata {
    name      = "documents-service-dashboard"
    namespace = "monitoring"
    labels    = { grafana_dashboard = "1" }
  }

  # Ruta relativa desde k8s/terraform/aws
  data = {
    "documents-service-dashboard.json" = file("${path.module}/../../../grafana/provisioning/dashboards/documents-service-dashboard.json")
  }
}

# -------- PrometheusRules for Alerts --------
# Define custom alerts for this microservice
resource "kubernetes_manifest" "prometheus_rule" {
  provider = kubernetes.eks
  manifest = {
    apiVersion = "monitoring.coreos.com/v1"
    kind       = "PrometheusRule"
    metadata = {
      name      = "documents-service-alerts"
      namespace = "monitoring"
      labels = {
        release = "kube-prometheus-stack"
      }
    }
    spec = yamldecode(file("${path.module}/../../../prometheus/alerts.yml"))
  }
}
