# -------- AWS Load Balancer Controller --------
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

  # Sintaxis moderna: set = [ {name="", value=""}, ... ]
  set = [
    { name = "clusterName",           value = var.cluster_name },
    { name = "serviceAccount.create", value = "true" },
    { name = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn", value = var.aws_lb_controller_role_arn }
  ]
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
  set = [
    { name = "installCRDs", value = "true" },
    { name = "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn", value = var.eso_irsa_role_arn }
  ]
  depends_on = [time_sleep.wait_for_alb_webhook]
}

// Removed standalone ServiceAccount to avoid conflicts; Helm creates it with IRSA annotation

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

# Espera breve para que el webhook/endpoints del ALB Controller estén disponibles
resource "time_sleep" "wait_for_alb_webhook" {
  depends_on      = [helm_release.aws_load_balancer_controller]
  create_duration = "45s"
}

# -------- Outputs útiles --------
output "kube_prom_stack_status" {
  value = helm_release.kube_prometheus_stack.status
}
