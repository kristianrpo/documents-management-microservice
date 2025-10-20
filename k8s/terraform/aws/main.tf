# ============================================================================
# Microservice-specific K8s resources
# Note: AWS Load Balancer Controller and External Secrets Operator are now
# deployed by the shared infrastructure repository
# ============================================================================

# -------- Grafana Dashboard (ConfigMap) --------
resource "kubernetes_namespace" "monitoring" {
  provider = kubernetes.eks
  metadata { name = "monitoring" }
}

resource "kubernetes_config_map" "grafana_dashboard" {
  provider = kubernetes.eks
  depends_on = [kubernetes_namespace.monitoring]
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

# -------- kube-prometheus-stack (Prometheus + Grafana) --------
resource "helm_release" "kube_prometheus_stack" {
  provider         = helm.eks
  name             = "kube-prometheus-stack"
  repository       = "https://prometheus-community.github.io/helm-charts"
  chart            = "kube-prometheus-stack"
  namespace        = "monitoring"
  create_namespace = true
  version          = "56.6.2"
  wait             = false
  atomic           = false
  timeout          = 600
  skip_crds        = false

  set = [
    { name = "prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues", value = "false" },
    { name = "grafana.service.type",                                             value = "LoadBalancer" },
    { name = "grafana.sidecar.dashboards.enabled",                               value = "true" },
    { name = "grafana.sidecar.dashboards.label",                                 value = "grafana_dashboard" }
  ]

  set_sensitive = [
    { name = "grafana.adminPassword", value = "admin" }
  ]

  # Inyectamos las reglas como parte del chart para evitar la carrera con la CRD
  values = [
    yamlencode({
      prometheus = {
        additionalPrometheusRulesMap = {
          "documents-service-alerts" = yamldecode(file("${path.module}/../../../prometheus/alerts.yml"))
        }
      }
    })
  ]

  depends_on = [kubernetes_namespace.monitoring]
}

# Espera no bloqueante del release pero con pausa para que CRDs y pods arranquen
resource "time_sleep" "wait_after_kps_install" {
  depends_on      = [helm_release.kube_prometheus_stack]
  create_duration = "60s"
}

# -------- Outputs útiles --------
output "kube_prom_stack_status" {
  value = helm_release.kube_prometheus_stack.status
}
