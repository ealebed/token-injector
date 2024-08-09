variable "project_id" {
  description = "Deployment Project ID"
  type        = string
}

variable "aws_role_arn" {
  description = "Amazon Resource Name (ARN) for role with configured acces to AWS, e.g. 'arn:aws:iam::123456789012:role/gcp_reader_role'"
  type        = string
  default     = "arn:aws:iam::123456789012:role/gcp_reader_role"
}

variable "certificator_image" {
  description = "Container image for creating K8S Secret (with TLS type) which contains private key and signed by K8S CA client certificate"
  type        = string
}

variable "webhook_image" {
  description = "Container image for mutating admission webhook"
  type        = string
}

variable "token_requester_image" {
  description = "Container image for getting Google Cloud ID token when running with under GCP Service Account"
  type        = string
}
