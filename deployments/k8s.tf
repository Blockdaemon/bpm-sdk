locals {
  common_labels = {
    "app.kubernetes.io/managed-by" = "terraform"
    "app.kubernetes.io/part-of"    = "bpm"
    "app.kubernetes.io/component"  = "sdk-docs"
    "app.kubernetes.io/name"       = "redoc"
  }
}

resource "kubernetes_namespace" "bpm_sdk_namespace" {
  metadata {
    name   = var.k8s_namespace
    labels = local.common_labels
  }
}

# Namespace limits according to https://blockdaemon.atlassian.net/wiki/spaces/BD/pages/1031241732/Kubernetes+Best+Practices
resource "kubernetes_limit_range" "bpm_sdk_namespace_limits" {
  metadata {
    name      = "${var.k8s_namespace}-limits"
    namespace = var.k8s_namespace
    labels    = local.common_labels
  }
  spec {
    limit {
      type = "Container"
      default = {
        cpu    = "50m"
        memory = "256Mi"
      }
    }
  }
}

# TODO: Move images to gitlab and replace secret
resource "kubernetes_secret" "bpm_sdk_image_pull_secret" {
  metadata {
    name      = var.k8s_image_pull_secret_name
    namespace = var.k8s_namespace
    labels    = local.common_labels
  }

  data = {
    ".dockerconfigjson" = "${file("${path.root}/secrets/dockerhub-regcred.json")}"
  }

  type = "kubernetes.io/dockerconfigjson"
}


resource "kubernetes_deployment" "bpm_sdk_deployment" {
  metadata {
    name      = "bpm-sdk-docs"
    namespace = var.k8s_namespace
    labels    = local.common_labels
  }

  spec {
    replicas = 1

    selector {
      match_labels = local.common_labels
    }

    template {
      metadata {
        labels = local.common_labels
      }

      spec {
        image_pull_secrets {
          name = var.k8s_image_pull_secret_name
        }

        container {
          image = var.redoc_image
          name  = "bpm-sdk-redoc"

          env {
            name  = "SPEC_URL"
            value = "https://${var.gcp_bucket_name}.storage.googleapis.com/swagger.yaml"
          }

          resources {
            limits {
              cpu    = "10m"
              memory = "50Mi"
            }
            requests {
              cpu    = "10m"
              memory = "50Mi"
            }
          }

          liveness_probe {
            http_get {
              path = "/"
              port = 80
            }

            initial_delay_seconds = 3
            period_seconds        = 3
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "bpm_sdk_svc" {
  metadata {
    namespace = var.k8s_namespace
    name      = "bpm-sdk-docs-svc"
    labels    = local.common_labels
  }
  spec {
    selector = local.common_labels
    type     = "ClusterIP"
    port {
      port        = 80
      target_port = 80
    }
  }
}

# Ensure minimum 1 pod is always available (e.g. during upgrades)
resource "kubernetes_pod_disruption_budget" "bpm_sdk_pdb" {
  metadata {
    namespace = var.k8s_namespace
    name      = "bpm-sdk-docs-pdb"
    labels    = local.common_labels
  }
  spec {
    selector {
      match_labels = local.common_labels
    }
    min_available = 1
  }
}

resource "kubernetes_ingress" "bpm_sdk_ingress" {
  metadata {
    namespace = var.k8s_namespace
    name      = "bpm-sdk-docs-ingress"
    labels    = local.common_labels

    annotations = {
      "kubernetes.io/ingress.class"       = "nginx"
      "certmanager.k8s.io/cluster-issuer" = "letsencrypt-prod"
      "kubernetes.io/tls-acme"            = "true"
    }
  }
  spec {
    tls {
      secret_name = "bpm-sdk-docs-tls-secret"
      hosts       = [local.fqdn]
    }
    rule {
      host = local.fqdn
      http {
        path {
          path = "/"
          backend {
            service_name = kubernetes_service.bpm_sdk_svc.metadata[0].name
            service_port = "80"
          }
        }
      }
    }
  }
}

