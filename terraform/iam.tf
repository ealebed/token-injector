module "aws_accessor_sa" {
  source  = "terraform-google-modules/service-accounts/google//modules/simple-sa"
  version = "~> 4.0"

  project_id  = var.project_id
  name        = "aws-accessor-sa"
  description = "Service Account used for accessing AWS resources from GKE"
  project_roles = [
    "roles/iam.serviceAccountTokenCreator",
  ]
}

resource "google_service_account_iam_member" "aws_accessor_sa_wiu" {
  service_account_id = module.aws_accessor_sa.id
  role               = "roles/iam.workloadIdentityUser"
  member             = "serviceAccount:${var.project_id}.svc.id.goog[application-namespace/aws-reader-sa]"

  depends_on = [module.aws_accessor_sa]
}
