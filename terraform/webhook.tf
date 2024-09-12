# GKE Service Account in relevant GKE namespace to run Mutating Admission webhook from
resource "kubernetes_service_account_v1" "webhook_sa" {
  metadata {
    name      = "webhook-sa-name"
    namespace = kubernetes_namespace_v1.webhook.metadata[0].name
    labels = {
      app = "admission-webhook"
    }
  }

  depends_on = [kubernetes_namespace_v1.webhook]
}

# GKE Service Account in relevant GKE namespace to create K8S Secret with TLS type that includes
# corresponding client certificate signed by K8S CA and private key
resource "kubernetes_service_account_v1" "webhook_cert_sa" {
  metadata {
    name      = "webhook-cert-sa"
    namespace = kubernetes_namespace_v1.webhook.metadata[0].name
    labels = {
      app = "admission-webhook"
    }
  }

  depends_on = [kubernetes_namespace_v1.webhook]
}

# Cluster Role for Mutating Admission webhook
resource "kubernetes_cluster_role_v1" "webhook_cr" {
  metadata {
    name = "admission-webhook-cr"
    labels = {
      app = "admission-webhook"
    }
  }
  rule {
    api_groups = [""]
    resources  = ["pods", "events"]
    verbs      = ["VerbAll"]
  }
  rule {
    api_groups = ["apps"]
    resources  = ["deployments", "daemonsets", "replicasets", "statefulsets"]
    verbs      = ["VerbAll"]
  }
  rule {
    api_groups = ["autoscaling"]
    resources  = ["horizontalpodautoscalers", "horizontalpodautoscalers/status"]
    verbs      = ["get", "list", "watch", "create", "patch", "update", "delete", "deletecollection"]
  }
  rule {
    api_groups = [""]
    resources  = ["serviceaccounts"]
    verbs      = ["get"]
  }
}

# Binding Cluster Role for Mutating Admission webhook to relevant GKE Service Account
resource "kubernetes_cluster_role_binding_v1" "webhook_crb" {
  metadata {
    name = "admission-webhook-crb"
    labels = {
      app = "admission-webhook"
    }
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role_v1.webhook_cr.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account_v1.webhook_sa.metadata[0].name
    namespace = kubernetes_namespace_v1.webhook.metadata[0].name
  }

  depends_on = [
    kubernetes_namespace_v1.webhook,
    kubernetes_service_account_v1.webhook_sa,
    kubernetes_cluster_role_v1.webhook_cr
  ]
}

# Cluster Role for creating secrets with client certificate which is signed by K8S CA and private key
resource "kubernetes_cluster_role_v1" "webhook_cert_cr" {
  metadata {
    name = "webhook-cert-sa-cr"
    labels = {
      app = "admission-webhook"
    }
  }
  rule {
    api_groups = ["admissionregistration.k8s.io"]
    resources  = ["mutatingwebhookconfigurations"]
    verbs      = ["get", "create", "patch"]
  }
  rule {
    api_groups = ["certificates.k8s.io"]
    resources  = ["certificatesigningrequests"]
    verbs      = ["get", "create", "delete", "list", "watch"]
  }
  rule {
    api_groups = ["certificates.k8s.io"]
    resources  = ["certificatesigningrequests/approval"]
    verbs      = ["update"]
  }
  rule {
    api_groups = ["certificates.k8s.io"]
    resources  = ["signers"]
    verbs      = ["approve"]
  }
  rule {
    api_groups = [""]
    resources  = ["secrets"]
    verbs      = ["create", "get", "patch", "update"]
  }
  rule {
    api_groups = [""]
    resources  = ["configmaps"]
    verbs      = ["get"]
  }
}

# Binding Cluster Role for creating secrets to relevant GKE Service Account
resource "kubernetes_cluster_role_binding_v1" "webhook_cert_crb" {
  metadata {
    name = "webhook-cert-sa-crb"
    labels = {
      app = "admission-webhook"
    }
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role_v1.webhook_cert_cr.metadata[0].name
  }
  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account_v1.webhook_cert_sa.metadata[0].name
    namespace = kubernetes_namespace_v1.webhook.metadata[0].name
  }

  depends_on = [
    kubernetes_namespace_v1.webhook,
    kubernetes_service_account_v1.webhook_cert_sa,
    kubernetes_cluster_role_v1.webhook_cert_cr
  ]
}

# Job for creating secrets with client certificate which is signed by K8S CA and private key
resource "kubernetes_job_v1" "webhook_cert_setup" {
  metadata {
    name      = "admission-webhook-cert-setup"
    namespace = kubernetes_namespace_v1.webhook.metadata[0].name
    labels = {
      app = "admission-webhook"
    }
  }
  spec {
    template {
      metadata {}
      spec {
        service_account_name = kubernetes_service_account_v1.webhook_cert_sa.metadata[0].name
        container {
          name              = "webhook-cert-setup"
          image             = var.certificator_image
          image_pull_policy = "Always"
          args              = ["certify", "--service", "admission-webhook-svc", "--namespace", kubernetes_namespace_v1.webhook.metadata[0].name, "--secret", "webhook-certs"]
        }
        restart_policy = "Never"
      }
    }
    backoff_limit = 0
  }

  wait_for_completion = true

  timeouts {
    create = "2m"
    update = "2m"
  }
}

# Deployment for admission webhook server
resource "kubernetes_deployment_v1" "webhook_deployment" {
  metadata {
    name      = "admission-webhook-deployment"
    namespace = kubernetes_namespace_v1.webhook.metadata[0].name
    labels = {
      app = "admission-webhook"
    }
  }
  spec {
    replicas = 1
    selector {
      match_labels = {
        app = "admission-webhook"
      }
    }
    template {
      metadata {
        labels = {
          app = "admission-webhook"
        }
      }
      spec {
        container {
          image             = var.webhook_image
          image_pull_policy = "Always"
          name              = "admission-webhook"
          args = [
            "--log-level=debug", "server",
            "--tls-cert-file=/etc/webhook/certs/tls.crt",
            "--tls-private-key-file=/etc/webhook/certs/tls.key",
            "--image=${var.token_requester_image}",
          "--pull-policy=Always"]
          security_context {
            read_only_root_filesystem = true
            capabilities {
              drop = ["NET_RAW", "ALL"]
            }
          }
          resources {
            requests = {
              cpu    = "250m"
              memory = "512Mi"
            }
            limits = {
              cpu    = "250m"
              memory = "512Mi"
            }
          }
          port {
            container_port = 8443
            name           = "https"
            protocol       = "TCP"
          }
          liveness_probe {
            http_get {
              path   = "/healthz"
              port   = "https"
              scheme = "HTTPS"
            }
            initial_delay_seconds = 3
            period_seconds        = 3
          }
          readiness_probe {
            http_get {
              path   = "/healthz"
              port   = "https"
              scheme = "HTTPS"
            }
            initial_delay_seconds = 3
            period_seconds        = 3
          }
          volume_mount {
            name       = "webhook-certs"
            mount_path = "/etc/webhook/certs"
            read_only  = true
          }
        }
        service_account_name = kubernetes_service_account_v1.webhook_sa.metadata[0].name
        volume {
          name = "webhook-certs"
          secret {
            secret_name = "webhook-certs"
          }
        }
      }
    }
  }

  depends_on = [kubernetes_job_v1.webhook_cert_setup]
}

# Service for admission webhook
resource "kubernetes_service_v1" "webhook_service" {
  metadata {
    name      = "admission-webhook-svc"
    namespace = kubernetes_namespace_v1.webhook.metadata[0].name
    labels = {
      app = "admission-webhook"
    }
  }
  spec {
    selector = {
      app = "admission-webhook"
    }
    port {
      port        = 443
      target_port = 8443
    }
  }

  depends_on = [kubernetes_deployment_v1.webhook_deployment]
}

# Mutating Webhook Configuration for admission webhook
resource "kubernetes_mutating_webhook_configuration_v1" "this" {
  metadata {
    name = "mutating-admission-webhook-cfg"
    labels = {
      app = "admission-webhook"
    }
  }
  webhook {
    name                      = "admission-webhook.example.com"
    admission_review_versions = ["v1", "v1beta1"]
    client_config {
      service {
        name      = kubernetes_service_v1.webhook_service.metadata[0].name
        namespace = kubernetes_namespace_v1.webhook.metadata[0].name
        path      = "/pods"
      }
      ca_bundle = base64decode(module.gke.ca_certificate)
    }
    object_selector {
      match_expressions {
        key      = "admission.token-injector/enabled"
        operator = "Exists"
      }
    }
    rule {
      api_groups   = ["*"]
      api_versions = ["*"]
      operations   = ["CREATE"]
      resources    = ["pods"]
      scope        = "Namespaced"
    }

    failure_policy = "Ignore"
    side_effects   = "None"
  }
}
