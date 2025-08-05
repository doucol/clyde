package util

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestParseImageVersion(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected string
	}{
		{
			name:     "Calico image with version",
			image:    "calico/node:v3.30.0",
			expected: "3.30.0",
		},
		{
			name:     "Tigera operator image",
			image:    "quay.io/tigera/operator:v1.32.0",
			expected: "1.32.0",
		},
		{
			name:     "Image with version without v prefix",
			image:    "calico/cni:3.28.1",
			expected: "3.28.1",
		},
		{
			name:     "Image with latest tag",
			image:    "calico/node:latest",
			expected: "latest",
		},
		{
			name:     "Image without tag",
			image:    "calico/node",
			expected: "",
		},
		{
			name:     "Empty image",
			image:    "",
			expected: "",
		},
		{
			name:     "Complex image path with version",
			image:    "registry.example.com/calico/node:v3.30.1",
			expected: "3.30.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseImageVersion(tt.image)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{
			name:     "v1 greater than v2",
			v1:       "3.30.0",
			v2:       "3.29.0",
			expected: true,
		},
		{
			name:     "v1 equal to v2",
			v1:       "3.30.0",
			v2:       "3.30.0",
			expected: true,
		},
		{
			name:     "v1 less than v2",
			v1:       "3.29.0",
			v2:       "3.30.0",
			expected: false,
		},
		{
			name:     "Major version difference",
			v1:       "4.0.0",
			v2:       "3.30.0",
			expected: true,
		},
		{
			name:     "Minor version difference",
			v1:       "3.30.1",
			v2:       "3.30.0",
			expected: true,
		},
		{
			name:     "Patch version equal",
			v1:       "3.30.1",
			v2:       "3.30.1",
			expected: true,
		},
		{
			name:     "Single digit versions",
			v1:       "1.0.0",
			v2:       "1.0.0",
			expected: true,
		},
		{
			name:     "Invalid version format v1",
			v1:       "invalid",
			v2:       "3.30.0",
			expected: false,
		},
		{
			name:     "Invalid version format v2",
			v1:       "3.30.0",
			v2:       "invalid",
			expected: true,
		},
		{
			name:     "Empty versions",
			v1:       "",
			v2:       "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SemverGreaterThanOrEqual(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("CompareVersions(%s, %s) = %v, expected %v", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestDetectCNIType(t *testing.T) {
	tests := []struct {
		name        string
		daemonSets  []appsv1.DaemonSet
		pods        []corev1.Pod
		expected    string
		expectError bool
	}{
		{
			name: "Calico DaemonSet in kube-system",
			daemonSets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-node",
						Namespace: "kube-system",
					},
				},
			},
			expected:    "Calico",
			expectError: false,
		},
		{
			name: "Calico DaemonSet in calico-system",
			daemonSets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-node",
						Namespace: "calico-system",
					},
				},
			},
			expected:    "Calico",
			expectError: false,
		},
		{
			name: "Flannel DaemonSet",
			daemonSets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kube-flannel-ds",
						Namespace: "kube-system",
					},
				},
			},
			expected:    "Flannel",
			expectError: false,
		},
		{
			name: "Cilium DaemonSet",
			daemonSets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cilium",
						Namespace: "kube-system",
					},
				},
			},
			expected:    "Cilium",
			expectError: false,
		},
		{
			name: "Calico pod with image in calico-system",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-node-abc123",
						Namespace: "calico-system",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "calico-node",
								Image: "calico/node:v3.30.0",
							},
						},
					},
				},
			},
			expected:    "Calico",
			expectError: false,
		},
		{
			name: "Unknown CNI - no matching DaemonSets or pods",
			daemonSets: []appsv1.DaemonSet{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "unknown-network",
						Namespace: "kube-system",
					},
				},
			},
			expected:    "Unknown",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create objects for fake client
			var objects []runtime.Object
			for _, ds := range tt.daemonSets {
				objects = append(objects, ds.DeepCopy())
			}
			for _, pod := range tt.pods {
				objects = append(objects, pod.DeepCopy())
			}

			client := fake.NewSimpleClientset(objects...)
			ctx := context.Background()

			result, err := DetectCNIType(ctx, client)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected CNI type %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetCalicoPods(t *testing.T) {
	tests := []struct {
		name            string
		pods            []corev1.Pod
		namespace       string
		expectedPods    int
		expectedVersion string
		expectError     bool
	}{
		{
			name: "Calico pods with versions",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-node-abc123",
						Namespace: "calico-system",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "calico-node",
								Image: "calico/node:v3.30.0",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-typha-def456",
						Namespace: "calico-system",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "calico-typha",
								Image: "calico/typha:v3.30.0",
							},
						},
					},
				},
			},
			namespace:       "calico-system",
			expectedPods:    2,
			expectedVersion: "3.30.0",
			expectError:     false,
		},
		{
			name: "Mixed versions - should return highest",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-node-old",
						Namespace: "calico-system",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "calico-node",
								Image: "calico/node:v3.29.0",
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-node-new",
						Namespace: "calico-system",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "calico-node",
								Image: "calico/node:v3.30.1",
							},
						},
					},
				},
			},
			namespace:       "calico-system",
			expectedPods:    2,
			expectedVersion: "3.30.1",
			expectError:     false,
		},
		{
			name:            "No pods in namespace",
			pods:            []corev1.Pod{},
			namespace:       "calico-system",
			expectedPods:    0,
			expectedVersion: "",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objects []runtime.Object
			for _, pod := range tt.pods {
				objects = append(objects, pod.DeepCopy())
			}

			client := fake.NewSimpleClientset(objects...)
			ctx := context.Background()

			pods, version, err := GetCalicoPods(ctx, client, tt.namespace)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(pods) != tt.expectedPods {
				t.Errorf("Expected %d pods, got %d", tt.expectedPods, len(pods))
			}

			if version != tt.expectedVersion {
				t.Errorf("Expected version %s, got %s", tt.expectedVersion, version)
			}
		})
	}
}

func TestGetOperatorPods(t *testing.T) {
	tests := []struct {
		name            string
		pods            []corev1.Pod
		deployments     []appsv1.Deployment
		namespace       string
		expectedPods    int
		expectedVersion string
		expectError     bool
	}{
		{
			name: "Tigera operator pod",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tigera-operator-abc123",
						Namespace: "tigera-operator",
						Labels: map[string]string{
							"k8s-app": "tigera-operator",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "tigera-operator",
								Image: "quay.io/tigera/operator:v1.32.0",
							},
						},
					},
				},
			},
			namespace:       "",
			expectedPods:    1,
			expectedVersion: "1.32.0",
			expectError:     false,
		},
		{
			name: "Operator deployment",
			deployments: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tigera-operator",
						Namespace: "tigera-operator",
					},
					Spec: appsv1.DeploymentSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "tigera-operator",
										Image: "quay.io/tigera/operator:v1.32.1",
									},
								},
							},
						},
					},
				},
			},
			namespace:       "",
			expectedPods:    1,
			expectedVersion: "1.32.1",
			expectError:     false,
		},
		{
			name:            "No operator found",
			pods:            []corev1.Pod{},
			deployments:     []appsv1.Deployment{},
			namespace:       "calico-system",
			expectedPods:    0,
			expectedVersion: "",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objects []runtime.Object
			for _, pod := range tt.pods {
				objects = append(objects, pod.DeepCopy())
			}
			for _, deploy := range tt.deployments {
				objects = append(objects, deploy.DeepCopy())
			}

			client := fake.NewSimpleClientset(objects...)
			ctx := context.Background()

			pods, version, err := GetOperatorPods(ctx, client, tt.namespace)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(pods) != tt.expectedPods {
				t.Errorf("Expected %d pods, got %d", tt.expectedPods, len(pods))
			}

			if version != tt.expectedVersion {
				t.Errorf("Expected version %s, got %s", tt.expectedVersion, version)
			}
		})
	}
}

func TestGetWhiskerAvailability(t *testing.T) {
	tests := []struct {
		name      string
		pods      []corev1.Pod
		namespace string
		expected  bool
	}{
		{
			name: "Whisker backend available and running",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-node-abc123",
						Namespace: "calico-system",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "whisker-backend",
								Image: "calico/whisker:v3.30.0",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			namespace: "calico-system",
			expected:  true,
		},
		{
			name: "Whisker backend exists but not running",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-node-abc123",
						Namespace: "calico-system",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "whisker-backend",
								Image: "calico/whisker:v3.30.0",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				},
			},
			namespace: "calico-system",
			expected:  false,
		},
		{
			name: "No whisker backend container",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "calico-node-abc123",
						Namespace: "calico-system",
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "calico-node",
								Image: "calico/node:v3.30.0",
							},
						},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
			},
			namespace: "calico-system",
			expected:  false,
		},
		{
			name:      "No pods in namespace",
			pods:      []corev1.Pod{},
			namespace: "calico-system",
			expected:  false,
		},
		{
			name:      "Empty namespace defaults to calico-system",
			pods:      []corev1.Pod{},
			namespace: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objects []runtime.Object
			for _, pod := range tt.pods {
				objects = append(objects, pod.DeepCopy())
			}

			client := fake.NewSimpleClientset(objects...)
			ctx := context.Background()

			result := GetWhiskerAvailability(ctx, client, tt.namespace)

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetServiceCIDRs(t *testing.T) {
	tests := []struct {
		name        string
		configMaps  []corev1.ConfigMap
		expected    []string
		expectError bool
	}{
		{
			name: "kubeadm-config with serviceSubnet",
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubeadm-config",
						Namespace: "kube-system",
					},
					Data: map[string]string{
						"ClusterConfiguration": `
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
networking:
  serviceSubnet: 10.96.0.0/12
  podSubnet: 10.244.0.0/16
`,
					},
				},
			},
			expected:    []string{"10.96.0.0/12"},
			expectError: false,
		},
		{
			name: "kubeadm-config with multiple service subnets",
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubeadm-config",
						Namespace: "kube-system",
					},
					Data: map[string]string{
						"ClusterConfiguration": `
networking:
  serviceSubnet: 10.96.0.0/12,fd00:10:96::/108
`,
					},
				},
			},
			expected:    []string{"10.96.0.0/12", "fd00:10:96::/108"},
			expectError: false,
		},
		{
			name:        "No kubeadm-config ConfigMap",
			configMaps:  []corev1.ConfigMap{},
			expected:    []string{},
			expectError: false,
		},
		{
			name: "kubeadm-config without serviceSubnet",
			configMaps: []corev1.ConfigMap{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "kubeadm-config",
						Namespace: "kube-system",
					},
					Data: map[string]string{
						"ClusterConfiguration": `
apiVersion: kubeadm.k8s.io/v1beta3
kind: ClusterConfiguration
`,
					},
				},
			},
			expected:    []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var objects []runtime.Object
			for _, cm := range tt.configMaps {
				objects = append(objects, cm.DeepCopy())
			}

			client := fake.NewSimpleClientset(objects...)
			ctx := context.Background()

			result, err := GetServiceCIDRs(ctx, client)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d CIDRs, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected CIDR %s at index %d, got %s", expected, i, result[i])
				}
			}
		})
	}
}

func TestClusterNetworkingInfo(t *testing.T) {
	// Test that ClusterNetworkingInfo struct is properly initialized
	info := ClusterNetworkingInfo{
		CNIType:           "Calico",
		CalicoInstalled:   true,
		CalicoVersion:     "3.30.0",
		OperatorInstalled: true,
		OperatorVersion:   "1.32.0",
		WhiskerAvailable:  true,
		PodCIDRs:          []string{"10.244.0.0/16"},
		ServiceCIDRs:      []string{"10.96.0.0/12"},
		Overlay:           "VXLAN",
		Encapsulation:     "Always",
	}

	if info.CNIType != "Calico" {
		t.Errorf("Expected CNI type Calico, got %s", info.CNIType)
	}

	if !info.CalicoInstalled {
		t.Errorf("Expected CalicoInstalled to be true")
	}

	if info.CalicoVersion != "3.30.0" {
		t.Errorf("Expected Calico version 3.30.0, got %s", info.CalicoVersion)
	}

	if !info.WhiskerAvailable {
		t.Errorf("Expected WhiskerAvailable to be true")
	}

	if len(info.PodCIDRs) != 1 || info.PodCIDRs[0] != "10.244.0.0/16" {
		t.Errorf("Expected PodCIDRs [10.244.0.0/16], got %v", info.PodCIDRs)
	}
}
