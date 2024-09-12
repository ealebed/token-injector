terraform {
  required_version = ">= 1.9.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~>5.41.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~>5.41.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.31.0"
    }
  }
}
