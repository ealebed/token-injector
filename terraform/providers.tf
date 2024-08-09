provider "google" {
  scopes = [
    "https://www.googleapis.com/auth/cloud-platform",
    "https://www.googleapis.com/auth/userinfo.email",
  ]
}

provider "google-beta" {
  scopes = [
    "https://www.googleapis.com/auth/cloud-platform",
    "https://www.googleapis.com/auth/userinfo.email",
  ]
}

data "google_client_config" "default" {}

provider "kubernetes" {
  host                   = "host"
  token                  = "token"
  cluster_ca_certificate = base64decode("kube_ca_certificate")
}
