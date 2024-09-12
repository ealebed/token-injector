variable "project_id" {
  description = "Deployment Project ID"
  type        = string
}

variable "region" {
  description = "Deployment region"
  type        = string
  default     = "us-central1"
}

variable "gke_node_machine_type" {
  description = "Machine type for GKE nodes"
  type        = string
  default     = "n2-standard-2"
}

variable "gke_release_channel" {
  description = "Release channel for GKE cluster"
  type        = string
  default     = "STABLE"
}

variable "certificator_image" {
  description = "Container image for creating K8S Secret (with TLS type) which contains private key and signed by K8S CA client certificate"
  type        = string
  default     = "ealebed/certificator:latest"
}

variable "webhook_image" {
  description = "Container image for mutating admission webhook"
  type        = string
  default     = "ealebed/token-injector-webhook:latest"
}

variable "token_requester_image" {
  description = "Container image for getting Google Cloud ID token when running with under GCP Service Account"
  type        = string
  default     = "ealebed/token-injector:latest"
}
