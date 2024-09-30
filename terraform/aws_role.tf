data "google_service_account" "aws_accessor_sa" {
  account_id = module.aws_accessor_sa.id
}

resource "aws_iam_role" "gcp_gke_role" {
  name = "gcp_gke_role"

  assume_role_policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
      {
        "Effect" = "Allow",
        "Principal" = {
          "Federated" = "accounts.google.com"
        },
        "Action" = "sts:AssumeRoleWithWebIdentity",
        "Condition" = {
          "StringEquals" = {
            "accounts.google.com:aud" = data.google_service_account.aws_accessor_sa.unique_id
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "gcp_gke_role_policy" {
  name = "gcp-gke-role-policy"
  role = aws_iam_role.gcp_gke_role.id

  # Terraform's "jsonencode" function converts a
  # Terraform expression result to valid JSON syntax.
  policy = jsonencode({
    "Version" = "2012-10-17",
    "Statement" = [
      {
        "Effect" = "Allow",
        "Action" = [
          "secretsmanager:GetResourcePolicy",
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret",
          "secretsmanager:ListSecretVersionIds",
          "secretsmanager:ListSecrets"
        ],
        "Resource" = "*"
      }
    ]
  })
}
