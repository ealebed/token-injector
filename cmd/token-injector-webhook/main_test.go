package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	fake "k8s.io/client-go/kubernetes/fake"
)

func TestMain(m *testing.M) {
	testMode = true
	os.Exit(m.Run())
}

//nolint:funlen
func Test_mutatingWebhook_mutateContainers(t *testing.T) {
	type fields struct {
		k8sClient  kubernetes.Interface
		image      string
		pullPolicy string
		volumeName string
		volumePath string
		tokenFile  string
	}
	type args struct {
		containers []corev1.Container
		roleArn    string
		ns         string
	}
	tests := []struct {
		name             string
		fields           fields
		args             args
		mutated          bool
		wantedContainers []corev1.Container
	}{
		{
			name: "mutate single container",
			fields: fields{
				k8sClient:  fake.NewSimpleClientset(),
				volumeName: "test-volume-name",
				volumePath: "/test-volume-path",
				tokenFile:  "test-token",
			},
			args: args{
				containers: []corev1.Container{
					{
						Name:  "TestContainer",
						Image: "test-image",
					},
				},
				roleArn: "arn:aws:iam::123456789012:role/testrole",
				ns:      "test-namespace",
			},
			wantedContainers: []corev1.Container{
				{
					Name:         "TestContainer",
					Image:        "test-image",
					VolumeMounts: []corev1.VolumeMount{{Name: "test-volume-name", MountPath: "/test-volume-path"}},
					Env: []corev1.EnvVar{
						{Name: awsWebIdentityTokenFile, Value: "/test-volume-path/test-token"},
						{Name: awsRoleArn, Value: "arn:aws:iam::123456789012:role/testrole"},
						{Name: awsRoleSessionName, Value: "token-injector-webhook-" + strings.Repeat("0", 16)},
					},
				},
			},
			mutated: true,
		},
		{
			name: "mutate multiple container",
			fields: fields{
				k8sClient:  fake.NewSimpleClientset(),
				volumeName: "test-volume-name",
				volumePath: "/test-volume-path",
				tokenFile:  "test-token",
			},
			args: args{
				containers: []corev1.Container{
					{
						Name:  "TestContainer1",
						Image: "test-image-1",
					},
					{
						Name:  "TestContainer2",
						Image: "test-image-2",
					},
				},
				roleArn: "arn:aws:iam::123456789012:role/testrole",
				ns:      "test-namespace",
			},
			wantedContainers: []corev1.Container{
				{
					Name:         "TestContainer1",
					Image:        "test-image-1",
					VolumeMounts: []corev1.VolumeMount{{Name: "test-volume-name", MountPath: "/test-volume-path"}},
					Env: []corev1.EnvVar{
						{Name: awsWebIdentityTokenFile, Value: "/test-volume-path/test-token"},
						{Name: awsRoleArn, Value: "arn:aws:iam::123456789012:role/testrole"},
						{Name: awsRoleSessionName, Value: "token-injector-webhook-" + strings.Repeat("0", 16)},
					},
				},
				{
					Name:         "TestContainer2",
					Image:        "test-image-2",
					VolumeMounts: []corev1.VolumeMount{{Name: "test-volume-name", MountPath: "/test-volume-path"}},
					Env: []corev1.EnvVar{
						{Name: awsWebIdentityTokenFile, Value: "/test-volume-path/test-token"},
						{Name: awsRoleArn, Value: "arn:aws:iam::123456789012:role/testrole"},
						{Name: awsRoleSessionName, Value: "token-injector-webhook-" + strings.Repeat("0", 16)},
					},
				},
			},
			mutated: true,
		},
		{
			name: "no containers to mutate",
			fields: fields{
				k8sClient:  fake.NewSimpleClientset(),
				volumeName: "test-volume-name",
				volumePath: "/test-volume-path",
			},
			args: args{
				roleArn: "arn:aws:iam::123456789012:role/testrole",
				ns:      "test-namespace",
			},
			mutated: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := &mutatingWebhook{
				k8sClient:  tt.fields.k8sClient,
				image:      tt.fields.image,
				pullPolicy: tt.fields.pullPolicy,
				volumeName: tt.fields.volumeName,
				volumePath: tt.fields.volumePath,
				tokenFile:  tt.fields.tokenFile,
			}
			got := mw.mutateContainers(tt.args.containers, tt.args.roleArn)
			if got != tt.mutated {
				t.Errorf("mutatingWebhook.mutateContainers() = %v, want %v", got, tt.mutated)
			}
			if !cmp.Equal(tt.args.containers, tt.wantedContainers) {
				t.Errorf("mutatingWebhook.mutateContainers() = diff %v", cmp.Diff(tt.args.containers, tt.wantedContainers))
			}
		})
	}
}

//nolint:funlen
func Test_mutatingWebhook_mutatePod(t *testing.T) {
	type fields struct {
		image      string
		pullPolicy string
		volumeName string
		volumePath string
		tokenFile  string
	}
	type args struct {
		pod                *corev1.Pod
		ns                 string
		serviceAccountName string
		annotations        map[string]string
		dryRun             bool
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		wantedPod *corev1.Pod
	}{
		{
			name: "mutate pod",
			fields: fields{
				image:      "ealebed/token-injector/token-injector:test",
				pullPolicy: "Always",
				volumeName: "test-volume-name",
				volumePath: "/test-volume-path",
				tokenFile:  "test-token",
			},
			args: args{
				pod: &corev1.Pod{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "TestContainer",
								Image: "test-image",
							},
						},
						ServiceAccountName: "test-sa",
					},
				},
				ns:                 "test-namespace",
				serviceAccountName: "test-sa",
				annotations:        map[string]string{awsRoleArnKey: "arn:aws:iam::123456789012:role/testrole"},
			},
			wantedPod: &corev1.Pod{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Name:    "generate-gcp-id-token",
							Image:   "ealebed/token-injector/token-injector:test",
							Command: []string{"/token-injector", "--file=/test-volume-path/test-token", "--refresh=false"},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(requestsCPU),
									corev1.ResourceMemory: resource.MustParse(requestsMemory),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(limitsCPU),
									corev1.ResourceMemory: resource.MustParse(limitsMemory),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test-volume-name",
									MountPath: "/test-volume-path",
								},
							},
							ImagePullPolicy: "Always",
						},
					},
					Containers: []corev1.Container{
						{
							Name:         "TestContainer",
							Image:        "test-image",
							VolumeMounts: []corev1.VolumeMount{{Name: "test-volume-name", MountPath: "/test-volume-path"}},
							Env: []corev1.EnvVar{
								{Name: awsWebIdentityTokenFile, Value: "/test-volume-path/test-token"},
								{Name: awsRoleArn, Value: "arn:aws:iam::123456789012:role/testrole"},
								{Name: awsRoleSessionName, Value: "token-injector-webhook-" + strings.Repeat("0", 16)},
							},
						},
						{
							Name:    "update-gcp-id-token",
							Image:   "ealebed/token-injector/token-injector:test",
							Command: []string{"/token-injector", "--file=/test-volume-path/test-token", "--refresh=true"},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(requestsCPU),
									corev1.ResourceMemory: resource.MustParse(requestsMemory),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(limitsCPU),
									corev1.ResourceMemory: resource.MustParse(limitsMemory),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "test-volume-name",
									MountPath: "/test-volume-path",
								},
							},
							ImagePullPolicy: "Always",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "test-volume-name",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
					},
					ServiceAccountName: "test-sa",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Name:        tt.args.serviceAccountName,
					Namespace:   tt.args.ns,
					Annotations: tt.args.annotations,
				},
			}
			mw := &mutatingWebhook{
				k8sClient:  fake.NewSimpleClientset(sa),
				image:      tt.fields.image,
				pullPolicy: tt.fields.pullPolicy,
				volumeName: tt.fields.volumeName,
				volumePath: tt.fields.volumePath,
				tokenFile:  tt.fields.tokenFile,
			}
			if err := mw.mutatePod(context.TODO(), tt.args.pod, tt.args.ns, tt.args.dryRun); (err != nil) != tt.wantErr {
				t.Errorf("mutatingWebhook.mutatePod() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !cmp.Equal(tt.args.pod, tt.wantedPod) {
				t.Errorf("mutatingWebhook.mutateContainers() = diff %v", cmp.Diff(tt.args.pod, tt.wantedPod))
			}
		})
	}
}

func Test_randomString(t *testing.T) {
	// Set test mode to ensure deterministic output
	originalTestMode := testMode
	testMode = true
	defer func() {
		testMode = originalTestMode
	}()

	tests := []struct {
		name string
		l    int
		want string
	}{
		{
			name: "length 0",
			l:    0,
			want: "",
		},
		{
			name: "length 1",
			l:    1,
			want: "0",
		},
		{
			name: "length 16 (default session name length)",
			l:    16,
			want: strings.Repeat("0", 16),
		},
		{
			name: "length 32",
			l:    32,
			want: strings.Repeat("0", 32),
		},
		{
			name: "length 100",
			l:    100,
			want: strings.Repeat("0", 100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := randomString(tt.l)
			if got != tt.want {
				t.Errorf("randomString() = %v, want %v", got, tt.want)
			}
			if len(got) != tt.l {
				t.Errorf("randomString() length = %v, want %v", len(got), tt.l)
			}
		})
	}

	// Test non-test mode (random output)
	testMode = false
	got1 := randomString(16)
	got2 := randomString(16)
	// In non-test mode, we can't predict the output, but we can verify:
	// 1. Length is correct
	if len(got1) != 16 {
		t.Errorf("randomString() length = %v, want 16", len(got1))
	}
	if len(got2) != 16 {
		t.Errorf("randomString() length = %v, want 16", len(got2))
	}
	// 2. Contains only lowercase letters
	for _, char := range got1 {
		if char < 'a' || char > 'z' {
			t.Errorf("randomString() contains invalid character: %c", char)
		}
	}
	// 3. Two calls should likely produce different results (very high probability)
	// Note: There's a tiny chance they could be the same, but it's negligible
	if got1 == got2 {
		t.Logf("Warning: randomString() produced same result twice (unlikely but possible)")
	}
}

func Test_getInjectorVolume(t *testing.T) {
	tests := []struct {
		name       string
		volumeName string
		want       corev1.Volume
	}{
		{
			name:       "default volume name",
			volumeName: tokenVolumeName,
			want: corev1.Volume{
				Name: tokenVolumeName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
		},
		{
			name:       "custom volume name",
			volumeName: "custom-token-volume",
			want: corev1.Volume{
				Name: "custom-token-volume",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
		},
		{
			name:       "empty volume name",
			volumeName: "",
			want: corev1.Volume{
				Name: "",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{
						Medium: corev1.StorageMediumMemory,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInjectorVolume(tt.volumeName)
			if got.Name != tt.want.Name {
				t.Errorf("getInjectorVolume() Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.VolumeSource.EmptyDir == nil {
				t.Errorf("getInjectorVolume() EmptyDir is nil")
			} else if got.VolumeSource.EmptyDir.Medium != tt.want.VolumeSource.EmptyDir.Medium {
				t.Errorf("getInjectorVolume() Medium = %v, want %v",
					got.VolumeSource.EmptyDir.Medium, tt.want.VolumeSource.EmptyDir.Medium)
			}
		})
	}
}

func Test_getInjectorContainer(t *testing.T) {
	// Set test mode for deterministic randomString output
	originalTestMode := testMode
	testMode = true
	defer func() {
		testMode = originalTestMode
	}()

	tests := []struct {
		name          string
		containerName string
		image         string
		pullPolicy    string
		volumeName    string
		volumePath    string
		tokenFile     string
		refresh       bool
		want          corev1.Container
	}{
		{
			name:          "init container without refresh",
			containerName: "generate-gcp-id-token",
			image:         "test-image:latest",
			pullPolicy:    "Always",
			volumeName:    "token-volume",
			volumePath:    "/var/run/secrets/aws/token",
			tokenFile:     "token",
			refresh:       false,
			want: corev1.Container{
				Name:            "generate-gcp-id-token",
				Image:           "test-image:latest",
				ImagePullPolicy: corev1.PullPolicy("Always"),
				Command: []string{
					"/token-injector",
					"--file=/var/run/secrets/aws/token/token",
					"--refresh=false",
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "token-volume",
						MountPath: "/var/run/secrets/aws/token",
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(requestsCPU),
						corev1.ResourceMemory: resource.MustParse(requestsMemory),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(limitsCPU),
						corev1.ResourceMemory: resource.MustParse(limitsMemory),
					},
				},
			},
		},
		{
			name:          "sidecar container with refresh",
			containerName: "update-gcp-id-token",
			image:         "test-image:v1.0.0",
			pullPolicy:    "IfNotPresent",
			volumeName:    tokenVolumeName,
			volumePath:    tokenVolumePath,
			tokenFile:     tokenFileName,
			refresh:       true,
			want: corev1.Container{
				Name:            "update-gcp-id-token",
				Image:           "test-image:v1.0.0",
				ImagePullPolicy: corev1.PullPolicy("IfNotPresent"),
				Command: []string{
					"/token-injector",
					"--file=/var/run/secrets/aws/token/token",
					"--refresh=true",
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      tokenVolumeName,
						MountPath: tokenVolumePath,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(requestsCPU),
						corev1.ResourceMemory: resource.MustParse(requestsMemory),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(limitsCPU),
						corev1.ResourceMemory: resource.MustParse(limitsMemory),
					},
				},
			},
		},
		{
			name:          "container with custom paths",
			containerName: "custom-container",
			image:         "custom-image:tag",
			pullPolicy:    "Never",
			volumeName:    "custom-volume",
			volumePath:    "/custom/path",
			tokenFile:     "custom-token",
			refresh:       true,
			want: corev1.Container{
				Name:            "custom-container",
				Image:           "custom-image:tag",
				ImagePullPolicy: corev1.PullPolicy("Never"),
				Command: []string{
					"/token-injector",
					"--file=/custom/path/custom-token",
					"--refresh=true",
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "custom-volume",
						MountPath: "/custom/path",
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(requestsCPU),
						corev1.ResourceMemory: resource.MustParse(requestsMemory),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse(limitsCPU),
						corev1.ResourceMemory: resource.MustParse(limitsMemory),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getInjectorContainer(tt.containerName, tt.image, tt.pullPolicy, tt.volumeName, tt.volumePath, tt.tokenFile, tt.refresh)

			// Check Name
			if got.Name != tt.want.Name {
				t.Errorf("getInjectorContainer() Name = %v, want %v", got.Name, tt.want.Name)
			}

			// Check Image
			if got.Image != tt.want.Image {
				t.Errorf("getInjectorContainer() Image = %v, want %v", got.Image, tt.want.Image)
			}

			// Check ImagePullPolicy
			if got.ImagePullPolicy != tt.want.ImagePullPolicy {
				t.Errorf("getInjectorContainer() ImagePullPolicy = %v, want %v", got.ImagePullPolicy, tt.want.ImagePullPolicy)
			}

			// Check Command
			if len(got.Command) != len(tt.want.Command) {
				t.Errorf("getInjectorContainer() Command length = %v, want %v", len(got.Command), len(tt.want.Command))
			} else {
				for i, cmd := range got.Command {
					if cmd != tt.want.Command[i] {
						t.Errorf("getInjectorContainer() Command[%d] = %v, want %v", i, cmd, tt.want.Command[i])
					}
				}
			}

			// Check VolumeMounts
			if len(got.VolumeMounts) != len(tt.want.VolumeMounts) {
				t.Errorf("getInjectorContainer() VolumeMounts length = %v, want %v", len(got.VolumeMounts), len(tt.want.VolumeMounts))
			} else {
				for i, vm := range got.VolumeMounts {
					if vm.Name != tt.want.VolumeMounts[i].Name {
						t.Errorf("getInjectorContainer() VolumeMounts[%d].Name = %v, want %v", i, vm.Name, tt.want.VolumeMounts[i].Name)
					}
					if vm.MountPath != tt.want.VolumeMounts[i].MountPath {
						t.Errorf("getInjectorContainer() VolumeMounts[%d].MountPath = %v, want %v", i, vm.MountPath, tt.want.VolumeMounts[i].MountPath)
					}
				}
			}

			// Check Resources
			if got.Resources.Requests.Cpu().Cmp(*tt.want.Resources.Requests.Cpu()) != 0 {
				t.Errorf("getInjectorContainer() Resources.Requests.CPU = %v, want %v",
					got.Resources.Requests.Cpu(), tt.want.Resources.Requests.Cpu())
			}
			if got.Resources.Requests.Memory().Cmp(*tt.want.Resources.Requests.Memory()) != 0 {
				t.Errorf("getInjectorContainer() Resources.Requests.Memory = %v, want %v",
					got.Resources.Requests.Memory(), tt.want.Resources.Requests.Memory())
			}
			if got.Resources.Limits.Cpu().Cmp(*tt.want.Resources.Limits.Cpu()) != 0 {
				t.Errorf("getInjectorContainer() Resources.Limits.CPU = %v, want %v",
					got.Resources.Limits.Cpu(), tt.want.Resources.Limits.Cpu())
			}
			if got.Resources.Limits.Memory().Cmp(*tt.want.Resources.Limits.Memory()) != 0 {
				t.Errorf("getInjectorContainer() Resources.Limits.Memory = %v, want %v",
					got.Resources.Limits.Memory(), tt.want.Resources.Limits.Memory())
			}
		})
	}
}

func Test_healthzHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: 200,
		},
		{
			name:           "POST request",
			method:         "POST",
			expectedStatus: 200,
		},
		{
			name:           "HEAD request",
			method:         "HEAD",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/healthz", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(healthzHandler)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("healthzHandler() status = %v, want %v", status, tt.expectedStatus)
			}
		})
	}
}
