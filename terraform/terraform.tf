terraform {
  required_version = ">= 1.9.5"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 6.4.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 6.4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.31.0"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.69"
    }
  }
}
