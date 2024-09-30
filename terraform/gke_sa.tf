# Creates a Kubernetes service account 'application-namespace' namespace.
# The service account is linked to a GCP service account (for GKE) and an AWS IAM role (for cross-cloud access).
resource "kubernetes_service_account_v1" "workload_sa" {
  metadata {
    name      = "aws-reader-sa"
    namespace = kubernetes_namespace_v1.workload.metadata[0].name
    annotations = {
      "iam.gke.io/gcp-service-account" = module.aws_accessor_sa.email
      "amazonaws.com/role-arn"         = aws_iam_role.gcp_gke_role.arn
    }
  }
 
  depends_on = [kubernetes_namespace_v1.workload]
}
