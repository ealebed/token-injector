# token-injector-webhook
The `token-injector-webhook` is a Kubernetes mutating admission webhook that mutates any k8s Pod running under **specially annotated Kubernetes Service Account** (see below) and labeled with
```yaml
admission.token-injector/enabled: "true"
```

## Kubernetes Service Account Annotations
To allow k8s Pod mutation process, the Pod need to be runned from a k8s Service Account annotated as shown below:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  annotations:
    amazonaws.com/role-arn: arn:aws:iam::${AWS_ACCOUNT_ID}:role/${AWS_ROLE_NAME}
    iam.gke.io/gcp-service-account: ${IAM_SA_NAME}@${IAM_SA_PROJECT_ID}.iam.gserviceaccount.com
  name: ${KSA_NAME}
  namespace: application-namespace
```

This could be done manually invoking commands:
```bash
kubectl annotate serviceaccount ${KSA_NAME} \
    --namespace ${NAMESPACE} \
    iam.gke.io/gcp-service-account=${IAM_SA_NAME}@${IAM_SA_PROJECT_ID}.iam.gserviceaccount.com
```
and
```bash
kubectl annotate serviceaccount ${KSA_NAME} \
    --namespace ${NAMESPACE} \
    amazonaws.com/role-arn=arn:aws:iam::${AWS_ACCOUNT_ID}:role/${AWS_ROLE_NAME}
```

Read [more](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity#kubernetes-sa-to-iam).

## Example k8s Pod Definition
Example k8s Pod definition which could be used for testing Kubernetes mutating admission webhook flow is described below:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: application-namespace
  labels:
    admission.token-injector/enabled: "true" # required
spec:
  serviceAccountName: ${KSA_NAME} # required
  containers:
  - name: test-pod
    image: mikesir87/aws-cli
    command: ["tail", "-f", "/dev/null"]
```
