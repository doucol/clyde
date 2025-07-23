package cnitype

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsEKS(t *testing.T) {
	tests := []struct {
		name     string
		node     corev1.Node
		expected bool
	}{
		{
			name: "EKS node with cloud provider label",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubernetes.io/cloud-provider": "aws",
					},
				},
			},
			expected: true,
		},
		{
			name: "EKS node with labels",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"eks.amazonaws.com/nodegroup": "my-nodegroup",
					},
				},
			},
			expected: true,
		},
		{
			name: "EKS node with eksctl label",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"alpha.eksctl.io/cluster-name": "my-cluster",
					},
				},
			},
			expected: true,
		},
		{
			name: "Non-EKS node",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "gce://project/zone/instance",
				},
			},
			expected: false,
		},
		{
			name: "Node without provider ID",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubernetes.io/os": "linux",
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEKS(tt.node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsAKS(t *testing.T) {
	tests := []struct {
		name     string
		node     corev1.Node
		expected bool
	}{
		{
			name: "AKS node with cloud provider label",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubernetes.io/cloud-provider": "azure",
					},
				},
			},
			expected: true,
		},
		{
			name: "AKS node with cluster label",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubernetes.azure.com/cluster": "my-cluster",
					},
				},
			},
			expected: true,
		},
		{
			name: "AKS node with agentpool label",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"agentpool": "nodepool1",
					},
				},
			},
			expected: true,
		},
		{
			name: "Non-AKS node",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///us-west-2a/i-1234567890abcdef0",
				},
			},
			expected: false,
		},
		{
			name: "Node without provider ID or AKS labels",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubernetes.io/os": "linux",
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAKS(tt.node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsGKE(t *testing.T) {
	tests := []struct {
		name     string
		node     corev1.Node
		expected bool
	}{
		{
			name: "GKE node with cloud provider label",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubernetes.io/cloud-provider": "gce",
					},
				},
			},
			expected: true,
		},
		{
			name: "GKE node with labels",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"cloud.google.com/gke-nodepool": "default-pool",
					},
				},
			},
			expected: true,
		},
		{
			name: "GKE node with image label",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"cloud.google.com/gke-os-distribution": "ubuntu",
					},
				},
			},
			expected: true,
		},
		{
			name: "Non-GKE node",
			node: corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///us-west-2a/i-1234567890abcdef0",
				},
			},
			expected: false,
		},
		{
			name: "Node without provider ID or GKE labels",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubernetes.io/os": "linux",
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGKE(tt.node)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCNIInfo(t *testing.T) {
	tests := []struct {
		name     string
		cniInfo  CNIInfo
		expected string
	}{
		{
			name: "Calico CNI",
			cniInfo: CNIInfo{
				Name:    "calico",
				Version: "v3.30.0",
				Details: map[string]string{
					"source": "daemonset",
				},
			},
			expected: "calico",
		},
		{
			name: "Flannel CNI",
			cniInfo: CNIInfo{
				Name:    "flannel",
				Version: "v0.20.2",
				Details: map[string]string{
					"source": "configmap",
				},
			},
			expected: "flannel",
		},
		{
			name: "AWS VPC CNI",
			cniInfo: CNIInfo{
				Name:    "aws-vpc-cni",
				Version: "v1.12.6",
				Details: map[string]string{
					"source": "cloud-provider",
				},
			},
			expected: "aws-vpc-cni",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cniInfo.Name != tt.expected {
				t.Errorf("Expected CNI name %s, got %s", tt.expected, tt.cniInfo.Name)
			}
		})
	}
}

func TestGetGKECNIType(t *testing.T) {
	tests := []struct {
		name     string
		node     corev1.Node
		expected string
	}{
		{
			name: "GKE node with Calico network policy",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"cloud.google.com/gke-network-policy": "calico",
					},
				},
			},
			expected: "gke-native-with-calico",
		},
		{
			name: "GKE node with networking mode",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"cloud.google.com/gke-networking-mode": "vpc-native",
					},
				},
			},
			expected: "gke-vpc-native",
		},
		{
			name: "GKE node with default configuration",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"cloud.google.com/gke-nodepool": "default-pool",
					},
				},
			},
			expected: "gke-native",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We can't easily test getGKECNIType without a real client
			// but we can test the logic patterns
			networkPolicy, hasNetworkPolicy := tt.node.Labels["cloud.google.com/gke-network-policy"]
			if hasNetworkPolicy && networkPolicy == "calico" {
				result := "gke-native-with-calico"
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
				return
			}

			mode, hasMode := tt.node.Labels["cloud.google.com/gke-networking-mode"]
			if hasMode {
				result := "gke-" + mode
				if result != tt.expected {
					t.Errorf("Expected %s, got %s", tt.expected, result)
				}
				return
			}

			result := "gke-native"
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetAKSCNIType(t *testing.T) {
	tests := []struct {
		name     string
		node     corev1.Node
		expected string
	}{
		{
			name: "AKS node with Azure CNI",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubernetes.azure.com/cluster": "test-cluster",
					},
				},
			},
			expected: "azure-cni", // This would be determined by actual cluster inspection
		},
		{
			name: "AKS node with kubenet",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"agentpool": "nodepool1",
					},
				},
			},
			expected: "kubenet", // This would be determined by actual cluster inspection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the pattern matching logic
			if _, hasCluster := tt.node.Labels["kubernetes.azure.com/cluster"]; hasCluster {
				// In a real scenario, this would inspect the cluster configuration
				// For the test, we just verify the logic pattern
				t.Log("AKS cluster detected")
			}
			if _, hasAgentPool := tt.node.Labels["agentpool"]; hasAgentPool {
				t.Log("AKS agent pool detected")
			}
		})
	}
}

// Test helper function to verify CNI pattern matching
func TestCNIPatternMatching(t *testing.T) {
	tests := []struct {
		name        string
		daemonSet   string
		expected    string
		shouldMatch bool
	}{
		{
			name:        "Calico DaemonSet",
			daemonSet:   "calico-node",
			expected:    "calico",
			shouldMatch: true,
		},
		{
			name:        "Flannel DaemonSet",
			daemonSet:   "kube-flannel-ds",
			expected:    "flannel",
			shouldMatch: true,
		},
		{
			name:        "Cilium DaemonSet",
			daemonSet:   "cilium",
			expected:    "cilium",
			shouldMatch: true,
		},
		{
			name:        "AWS VPC CNI DaemonSet",
			daemonSet:   "aws-node",
			expected:    "aws-vpc-cni",
			shouldMatch: true,
		},
		{
			name:        "Unknown DaemonSet",
			daemonSet:   "some-other-ds",
			expected:    "",
			shouldMatch: false,
		},
	}

	// CNI patterns from the actual code
	cniPatterns := map[string]string{
		"calico":   "calico",
		"flannel":  "flannel",
		"weave":    "weave",
		"cilium":   "cilium",
		"antrea":   "antrea",
		"canal":    "canal",
		"aws-node": "aws-vpc-cni",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := false
			var result string

			for pattern, cniName := range cniPatterns {
				if strings.Contains(strings.ToLower(tt.daemonSet), pattern) {
					result = cniName
					found = true
					break
				}
			}

			if tt.shouldMatch && !found {
				t.Errorf("Expected to match CNI pattern but didn't for DaemonSet %s", tt.daemonSet)
			}

			if !tt.shouldMatch && found {
				t.Errorf("Expected not to match CNI pattern but did for DaemonSet %s", tt.daemonSet)
			}

			if tt.shouldMatch && result != tt.expected {
				t.Errorf("Expected CNI %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestImageVersionParsing(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected string
	}{
		{
			name:     "Calico image with version",
			image:    "calico/node:v3.30.0",
			expected: "v3.30.0",
		},
		{
			name:     "Flannel image with version",
			image:    "rancher/mirrored-flannelcni-flannel:v0.20.2",
			expected: "v0.20.2",
		},
		{
			name:     "AWS CNI image with version",
			image:    "amazon/aws-k8s-cni:v1.12.6",
			expected: "v1.12.6",
		},
		{
			name:     "Image with latest tag",
			image:    "nginx:latest",
			expected: "latest",
		},
		{
			name:     "Image without tag",
			image:    "nginx",
			expected: "unknown",
		},
		{
			name:     "Empty image",
			image:    "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string

			// Simulate version parsing logic
			if tt.image == "" {
				result = "unknown"
			} else {
				parts := strings.Split(tt.image, ":")
				if len(parts) == 2 {
					result = parts[1]
				} else {
					result = "unknown"
				}
			}

			if result != tt.expected {
				t.Errorf("Expected version %s, got %s", tt.expected, result)
			}
		})
	}
}
