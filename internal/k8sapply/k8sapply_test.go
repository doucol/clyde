package k8sapply

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	faked "k8s.io/client-go/dynamic/fake"
	fakec "k8s.io/client-go/kubernetes/fake"
)

func newApplierForTest(t *testing.T) *Applier {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	fakeC := fakec.NewClientset()
	disco := fakeC.Discovery().(*fakediscovery.FakeDiscovery)
	disco.FakedServerVersion = &version.Info{
		Major:      "1",
		Minor:      "32",
		GitVersion: "v1.32.0",
	}
	disco.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod", Group: "", Version: "v1"},
				{Name: "services", Namespaced: true, Kind: "Service", Group: "", Version: "v1"},
				{Name: "deployments", Namespaced: true, Kind: "Deployment", Group: "apps", Version: "v1"},
				{Name: "configmaps", Namespaced: true, Kind: "ConfigMap", Group: "", Version: "v1"},
				{Name: "secrets", Namespaced: true, Kind: "Secret", Group: "", Version: "v1"},
				{Name: "namespaces", Namespaced: false, Kind: "Namespace", Group: "", Version: "v1"},
				{Name: "daemonsets", Namespaced: true, Kind: "DaemonSet", Group: "apps", Version: "v1"},
				{Name: "statefulsets", Namespaced: true, Kind: "StatefulSet", Group: "apps", Version: "v1"},
				{Name: "jobs", Namespaced: true, Kind: "Job", Group: "batch", Version: "v1"},
				{Name: "cronjobs", Namespaced: true, Kind: "CronJob", Group: "batch", Version: "v1"},
				{Name: "ingresses", Namespaced: true, Kind: "Ingress", Group: "networking.k8s.io", Version: "v1"},
				{Name: "persistentvolumeclaims", Namespaced: true, Kind: "PersistentVolumeClaim", Group: "", Version: "v1"},
				{Name: "persistentvolumes", Namespaced: false, Kind: "PersistentVolume", Group: "", Version: "v1"},
				{Name: "serviceaccounts", Namespaced: true, Kind: "ServiceAccount", Group: "", Version: "v1"},
				{Name: "roles", Namespaced: true, Kind: "Role", Group: "rbac.authorization.k8s.io", Version: "v1"},
				{Name: "rolebindings", Namespaced: true, Kind: "RoleBinding", Group: "rbac.authorization.k8s.io", Version: "v1"},
				{Name: "clusterroles", Namespaced: false, Kind: "ClusterRole", Group: "rbac.authorization.k8s.io", Version: "v1"},
				{Name: "clusterrolebindings", Namespaced: false, Kind: "ClusterRoleBinding", Group: "rbac.authorization.k8s.io", Version: "v1"},
				{Name: "networkpolicies", Namespaced: true, Kind: "NetworkPolicy", Group: "networking.k8s.io", Version: "v1"},
				{Name: "poddisruptionbudgets", Namespaced: true, Kind: "PodDisruptionBudget", Group: "policy", Version: "v1"},
				{Name: "horizontalpodautoscalers", Namespaced: true, Kind: "HorizontalPodAutoscaler", Group: "autoscaling", Version: "v2"},
				{Name: "limitranges", Namespaced: true, Kind: "LimitRange", Group: "", Version: "v1"},
				{Name: "resourcequotas", Namespaced: true, Kind: "ResourceQuota", Group: "", Version: "v1"},
				{Name: "endpoints", Namespaced: true, Kind: "Endpoints", Group: "", Version: "v1"},
				{Name: "events", Namespaced: true, Kind: "Event", Group: "events.k8s.io", Version: "v1"},
				{Name: "replicasets", Namespaced: true, Kind: "ReplicaSet", Group: "apps", Version: "v1"},
				{Name: "replicationcontrollers", Namespaced: true, Kind: "ReplicationController", Group: "", Version: "v1"},
			},
		},
	}
	fakeD := faked.NewSimpleDynamicClient(scheme)
	a, err := NewApplier(fakeC, fakeD, logrus.StandardLogger().Out)
	if err != nil {
		t.Fatalf("Failed to create Applier: %v", err)
		return nil
	}
	return a
}

// func TestGetResourceFromKind(t *testing.T) {
// 	tests := []struct {
// 		group    string
// 		version  string
// 		kind     string
// 		expected string
// 	}{
// 		{"", "v1", "Pod", "pods"},
// 		{"", "v1", "Service", "services"},
// 		{"", "v1", "ConfigMap", "configmaps"},
// 		{"", "v1", "Secret", "secrets"},
// 		{"", "v1", "Namespace", "namespaces"},
// 		{"", "v1", "PersistentVolumeClaim", "persistentvolumeclaims"},
// 		{"", "v1", "PersistentVolume", "persistentvolumes"},
// 		{"", "v1", "ServiceAccount", "serviceaccounts"},
// 		{"", "v1", "LimitRange", "limitranges"},
// 		{"", "v1", "ResourceQuota", "resourcequotas"},
// 		{"", "v1", "Endpoints", "endpoints"},
// 		{"", "v1", "ReplicationController", "replicationcontrollers"},
// 		{"apps", "v1", "Deployment", "deployments"},
// 		{"apps", "v1", "DaemonSet", "daemonsets"},
// 		{"apps", "v1", "StatefulSet", "statefulsets"},
// 		{"apps", "v1", "ReplicaSet", "replicasets"},
// 		{"autoscaling", "v2", "HorizontalPodAutoscaler", "horizontalpodautoscalers"},
// 		{"batch", "v1", "Job", "jobs"},
// 		{"batch", "v1", "CronJob", "cronjobs"},
// 		{"events.k8s.io", "v1", "Event", "events"},
// 		{"networking.k8s.io", "v1", "Ingress", "ingresses"},
// 		{"networking.k8s.io", "v1", "NetworkPolicy", "networkpolicies"},
// 		{"policy", "v1", "PodDisruptionBudget", "poddisruptionbudgets"},
// 		{"rbac.authorization.k8s.io", "v1", "Role", "roles"},
// 		{"rbac.authorization.k8s.io", "v1", "RoleBinding", "rolebindings"},
// 		{"rbac.authorization.k8s.io", "v1", "ClusterRole", "clusterroles"},
// 		{"rbac.authorization.k8s.io", "v1", "ClusterRoleBinding", "clusterrolebindings"},
// 	}
// 	applier := newApplierForTest(t)
//
// 	for _, tt := range tests {
// 		t.Run(tt.kind, func(t *testing.T) {
// 			var err error
// 			var gvr schema.GroupVersionResource
// 			gvk := schema.GroupVersionKind{Group: tt.group, Version: tt.version, Kind: tt.kind}
// 			if gvr, err = applier.findGVRForGVK(gvk); err != nil {
// 				t.Errorf("getResourceFromKind(%s) error: %v", tt.kind, err)
// 			} else {
// 				if gvr.Resource != tt.expected {
// 					t.Errorf("getResourceFromKind(%s) = %s, want %s", tt.kind, gvr.Resource, tt.expected)
// 				}
// 			}
// 		})
// 	}
// }

func TestApplyString(t *testing.T) {
	ctx := context.Background()
	applier := newApplierForTest(t)

	// Test with a simple ConfigMap
	yamlContent := `
 apiVersion: v1
 kind: ConfigMap
 metadata:
   name: test-configmap
   namespace: default
 data:
   key1: value1
   key2: value2
 `

	err := applier.ApplyString(ctx, yamlContent)
	if err != nil {
		t.Errorf("ApplyString failed: %v", err)
	}
}

func TestApplyStringMultipleResources(t *testing.T) {
	applier := newApplierForTest(t)
	ctx := context.Background()

	// Test with multiple resources
	yamlContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap-1
  namespace: default
data:
  key1: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap-2
  namespace: default
data:
  key2: value2
`

	err := applier.ApplyString(ctx, yamlContent)
	if err != nil {
		t.Errorf("ApplyString with multiple resources failed: %v", err)
	}
}

func TestApplyStringEmptyYAML(t *testing.T) {
	applier := newApplierForTest(t)

	ctx := context.Background()

	// Test with empty YAML
	yamlContent := ""

	err := applier.ApplyString(ctx, yamlContent)
	if err != nil {
		t.Errorf("ApplyString with empty YAML failed: %v", err)
	}
}

func TestApplyStringInvalidYAML(t *testing.T) {
	applier := newApplierForTest(t)

	ctx := context.Background()

	// Test with invalid YAML
	yamlContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap
  namespace: default
data:
  key1: value1
  key2: [invalid: yaml
`

	err := applier.ApplyString(ctx, yamlContent)
	if err == nil {
		t.Error("ApplyString with invalid YAML should have failed")
	}
}

func TestApplyStringWithCustomResource(t *testing.T) {
	applier := newApplierForTest(t)

	ctx := context.Background()

	// Test with a custom resource
	yamlContent := `
apiVersion: custom.example.com/v1
kind: CustomResource
metadata:
  name: test-custom-resource
  namespace: default
spec:
  field1: value1
`

	err := applier.ApplyString(ctx, yamlContent)
	// The custom resource should be handled gracefully even if the CRD doesn't exist
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "could not find") {
		t.Errorf("ApplyString with custom resource failed unexpectedly: %v", err)
	}
}

func TestApplyStringWithCustomResourceNoNamespace(t *testing.T) {
	applier := newApplierForTest(t)

	ctx := context.Background()

	// Test with a cluster-scoped custom resource
	yamlContent := `
apiVersion: custom.example.com/v1
kind: CustomClusterResource
metadata:
  name: test-custom-cluster-resource
spec:
  field1: value1
`

	err := applier.ApplyString(ctx, yamlContent)
	// The custom resource should be handled gracefully even if the CRD doesn't exist
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "could not find") {
		t.Errorf("ApplyString with cluster-scoped custom resource failed unexpectedly: %v", err)
	}
}

func TestApplyStringWithNonNamespacedResources(t *testing.T) {
	applier := newApplierForTest(t)
	ctx := context.Background()

	// Test with non-namespaced resources
	yamlContent := `
apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: test-pv
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: /tmp/test
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: test-cluster-role
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: test-cluster-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: test-cluster-role
subjects:
- kind: ServiceAccount
  name: default
  namespace: default
`

	err := applier.ApplyString(ctx, yamlContent)
	if err != nil {
		t.Errorf("ApplyString with non-namespaced resources failed: %v", err)
	}
}

func TestApplyStringWithMixedResources(t *testing.T) {
	applier := newApplierForTest(t)
	ctx := context.Background()

	// Test with both namespaced and non-namespaced resources
	yamlContent := `
apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace-mixed
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap-mixed
  namespace: default
data:
  key1: value1
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: test-cluster-role-mixed
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list"]
`

	err := applier.ApplyString(ctx, yamlContent)
	if err != nil {
		t.Errorf("ApplyString with mixed resources failed: %v", err)
	}
}

func TestApplyURL(t *testing.T) {
	applier := newApplierForTest(t)

	ctx := context.Background()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		yamlContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap-url
  namespace: default
data:
  key1: value1
  key2: value2
`
		w.Header().Set("Content-Type", "text/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	// Test applying from URL
	err := applier.ApplyURL(ctx, server.URL)
	if err != nil {
		t.Errorf("ApplyURL failed: %v", err)
	}
}

func TestApplyURLWithTimeout(t *testing.T) {
	// This test requires a running Kubernetes cluster
	// Skip if not available
	applier := newApplierForTest(t)

	ctx := context.Background()

	// Create a test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Small delay
		yamlContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap-timeout
  namespace: default
data:
  key1: value1
`
		w.Header().Set("Content-Type", "text/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	// Test applying from URL with custom timeout
	err := applier.ApplyURLWithTimeout(ctx, server.URL, 1*time.Second)
	if err != nil {
		t.Errorf("ApplyURLWithTimeout failed: %v", err)
	}
}

func TestApplyURLWithTimeoutExceeded(t *testing.T) {
	// This test requires a running Kubernetes cluster
	// Skip if not available
	applier := newApplierForTest(t)

	ctx := context.Background()

	// Create a test server that delays response longer than timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer delay than timeout
		yamlContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap-timeout-exceeded
  namespace: default
data:
  key1: value1
`
		w.Header().Set("Content-Type", "text/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	// Test applying from URL with short timeout (should fail)
	err := applier.ApplyURLWithTimeout(ctx, server.URL, 100*time.Millisecond)
	if err == nil {
		t.Error("ApplyURLWithTimeout should have failed due to timeout")
	}
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestApplyURLWithHTTPError(t *testing.T) {
	applier := newApplierForTest(t)

	ctx := context.Background()

	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	// Test applying from URL that returns 404
	err := applier.ApplyURL(ctx, server.URL)
	if err == nil {
		t.Error("ApplyURL should have failed with 404 error")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("Expected 404 error, got: %v", err)
	}
}

func TestApplyURLWithInvalidYAML(t *testing.T) {
	applier := newApplierForTest(t)

	ctx := context.Background()

	// Create a test server that returns invalid YAML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		yamlContent := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: test-configmap-invalid
  namespace: default
data:
  key1: value1
  key2: [invalid: yaml
`
		w.Header().Set("Content-Type", "text/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(yamlContent))
	}))
	defer server.Close()

	// Test applying from URL with invalid YAML
	err := applier.ApplyURL(ctx, server.URL)
	if err == nil {
		t.Error("ApplyURL should have failed with invalid YAML")
	}
}
