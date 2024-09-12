# Webhook Configuration Description

## General information and prerequisites
1. Container images for `token-injector-webhook`, `token-injector` tool and `certificator` tool (see separate [repository](https://github.com/ealebed/admission-webhook-certificator)) should be built and uploaded to the Container Registry accessible from GKE cluster.

2. Google Service account which will be used for accessing AWS resources must be created with the following roles:
  - `roles/iam.workloadIdentityUser` - to impersonate service accounts from GKE Workloads
  - `roles/iam.serviceAccountTokenCreator` - to create OAuth2 access tokens, sign blobs, or sign JWTs

3. The Google Service account OAuth 2 Client ID value should be used when configuring AWS IAM Role Trusted entities.
  - To obtain OAuth 2 Client ID from Google Cloud Console go to *IAM & Admin* -> *Service accounts* page, find the needed Service Account and copy the Client ID value, e.g. `115304415613939302199`
  - To obtain OAuth 2 Client ID from terminal use the following command:
  ```bash
  gcloud iam service-accounts describe --format json ${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com  | jq -r '.uniqueId'
  ```
  where `${GSA_NAME}` is the name of the GCP Service Account and `${PROJECT_ID}` is your Google Project ID.

4. The GKE cluster with enabled Workload Identity must be created.

5. AWS IAM Role with Google OIDC Web Identity must be created with attached permissions policy and trust policy documents. See [doc](aws_role_creation.md)

6. GKE namespaces for Application workload and Admission webhook.

7. GKE Service Account for creation secret (with signed certificate and private key) and running Admission webhook in respective GKE namespace (see example in separate [repository](https://github.com/ealebed/admission-webhook-certificator/blob/master/manifests/service-account.yaml)).

8. GKE Cluster Role and Cluster Role Binding for Admission webhook Service Account.

9. GKE Cluster Role and Cluster Role Binding for `certificator` tool Service Account (see exapmles in separate [repository](https://github.com/ealebed/admission-webhook-certificator/tree/master/manifests))

10. Annotated GKE Service Account for running Application workload in respective GKE namespace.

11. GKE Job for `certificator` tool (used for creation secret with signed certificate and private key) (see example in separate [repository](https://github.com/ealebed/admission-webhook-certificator/blob/master/manifests/job.yaml)).

12. GKE Deployment for running Admission webhook in respective GKE namespace.

13. GKE Service which points to the Admission webhook Deployment in the respective GKE namespace.

14. GKE MutatingWebhookConfiguration for registering Admission webhook. Read [more](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/).

## Configuration steps
All required steps for configuring Kubernetes Mutating Admission webhook are described in separate documents:
- [Manual configuration](./docs/manual_configuration.md)
- [Terraform configuration](./docs/terraform_configuration.md)
- [HELM chart configuration](./docs/helm_configuration.md)
