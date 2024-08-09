resource "kubernetes_namespace_v1" "app_workload" {
  metadata {
    name = "application-namespace"
  }
}

resource "kubernetes_service_account_v1" "app_workload_sa" {
  metadata {
    name      = "aws-reader-sa"
    namespace = kubernetes_namespace_v1.app_workload.metadata[0].name
    annotations = {
      "iam.gke.io/gcp-service-account" = "${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
      "amazonaws.com/role-arn"         = var.aws_role_arn
    }
  }

  depends_on = [kubernetes_namespace_v1.app_workload]
}
