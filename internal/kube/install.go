package kube

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Goldmane and Whisker are cluster-scoped operator custom resources. Creating
// them tells the Tigera/Calico operator to stand up the flow log aggregator
// (Goldmane) and its backend/UI (Whisker) on a Calico v3.30+ cluster.
var (
	goldmaneGVR = schema.GroupVersionResource{Group: "operator.tigera.io", Version: "v1", Resource: "goldmanes"}
	whiskerGVR  = schema.GroupVersionResource{Group: "operator.tigera.io", Version: "v1", Resource: "whiskers"}
)

// CanInstallWhisker reports whether Clyde can attempt to enable Goldmane/Whisker
// on the inspected cluster. Installation works by creating operator custom
// resources, so it requires the Tigera/Calico operator to be present to
// reconcile them.
func CanInstallWhisker(info ClusterNetworkingInfo) bool {
	return info.OperatorInstalled
}

// InstallGoldmaneWhisker creates the Goldmane and Whisker operator custom
// resources that enable flow observability. It is idempotent: resources that
// already exist are left untouched. A missing CRD (i.e. the cluster's Calico
// version predates Whisker support) is surfaced as a clear error.
func InstallGoldmaneWhisker(ctx context.Context, dyn dynamic.Interface) error {
	resources := []struct {
		gvr  schema.GroupVersionResource
		kind string
	}{
		{goldmaneGVR, "Goldmane"},
		{whiskerGVR, "Whisker"},
	}
	for _, r := range resources {
		obj := &unstructured.Unstructured{
			Object: map[string]any{
				"apiVersion": "operator.tigera.io/v1",
				"kind":       r.kind,
				"metadata": map[string]any{
					"name": "default",
				},
			},
		}
		_, err := dyn.Resource(r.gvr).Create(ctx, obj, metav1.CreateOptions{})
		if err == nil || apierrors.IsAlreadyExists(err) {
			continue
		}
		if apimeta.IsNoMatchError(err) || apierrors.IsNotFound(err) {
			return fmt.Errorf("this cluster's Calico version does not support Whisker (Calico v3.30+ is required): %w", err)
		}
		return fmt.Errorf("creating %s resource: %w", r.kind, err)
	}
	return nil
}

// WaitForWhiskerAvailable polls until the whisker-backend pod is running in the
// given namespace, or the context is cancelled or the timeout elapses. The
// operator can take a little while to reconcile the new resources, so callers
// should allow a generous timeout.
func WaitForWhiskerAvailable(ctx context.Context, clientset kubernetes.Interface, namespace string, timeout time.Duration) error {
	if namespace == "" {
		namespace = "calico-system"
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Check immediately, then on every tick.
	for {
		if GetWhiskerAvailability(ctx, clientset, namespace) {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for Whisker to become available in namespace %q: %w", namespace, ctx.Err())
		case <-ticker.C:
		}
	}
}
