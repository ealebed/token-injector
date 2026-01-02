# Webhook Configuration with Terraform

## Prerequisites
- Container images for `token-injector-webhook`, `token-injector` tool and `certificator` tool (see separate [repository](https://github.com/ealebed/admission-webhook-certificator)) should be built and uploaded to the Container Registry accessible from GKE cluster.
- Export AWS credentials into environment variables
```bash
export AWS_ACCESS_KEY_ID="aws-access-key-id"
export AWS_SECRET_ACCESS_KEY="aws-secret-access-key"
export AWS_REGION="us-east-1"
```
- Login to gcp
```bash
gcloud auth application-default login
```

### Deploy Terraform code
```bash
cd terraform
terraform init
terraform plan
terraform apply
```

### Configure `kubectl` command line access by running the following command
```bash
gcloud container clusters get-credentials ${GKE_CLUSTER_NAME} --region ${GCP_REGION} --project ${PROJECT_ID}
```
For example:
```bash
gcloud container clusters get-credentials regional-cluster-test --region us-west1 --project ylebi-rnd
```

## Testing
- Create a k8s Pod configuration as shown below:
```bash
cat > test-pod.yaml << EOF
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: "application-namespace"
  labels:
    admission.token-injector/enabled: "true"
spec:
  serviceAccountName: "aws-reader-sa"
  containers:
  - name: test-pod
    image: mikesir87/aws-cli
    command: ["tail", "-f", "/dev/null"]
EOF
```

- Apply pod configuration from YAML file:
```bash
kubectl apply -f test-pod.yaml
```

- Connect to the running k8s Pod:
```bash
kubectl exec -it test-pod -n application-namespace -- bash
```

- Run the following command in the Pod shell to check the AWS assumed role:
```bash
aws sts get-caller-identity
```

The output should look similar to the below:
```text
{
    "UserId": "AROAXXPBSFGLKUDFGHT7Q:token-injector-webhook-luwkompqhfewtygb",
    "Account": "531438381462",
    "Arn": "arn:aws:sts::531438381462:assumed-role/gke-reader-role/token-injector-webhook-luwkompqhfewtygb"
}
```
