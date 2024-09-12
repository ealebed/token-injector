# GKE namespace for running Mutating Admission webhook and Certificator tool
resource "kubernetes_namespace_v1" "webhook" {
  metadata {
    name = "webhook"
    labels = {
      app = "admission-webhook"
    }
  }
}

# GKE namespace for running application workload
resource "kubernetes_namespace_v1" "workload" {
  metadata {
    name = "application-namespace"
  }
}
