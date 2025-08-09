package calico

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// WaitForTigeraOperatorAvailable waits for all Tigera operator status resources to become available
func WaitForTigeraOperatorAvailable(ctx context.Context, clientset *kubernetes.Clientset, dynamicClient dynamic.Interface, timeout time.Duration) error {
	// Define the GVR for TigeraStatus resource
	tigeraStatusGVR := schema.GroupVersionResource{
		Group:    "operator.tigera.io",
		Version:  "v1",
		Resource: "tigerastatuses",
	}

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Check initial state first
	if available, err := checkAllTigeraStatusesAvailable(ctxWithTimeout, dynamicClient, tigeraStatusGVR); err != nil {
		return fmt.Errorf("initial status check failed: %w", err)
	} else if available {
		return nil
	}

	// Start watching TigeraStatus resources
	watchInterface, err := dynamicClient.Resource(tigeraStatusGVR).Watch(ctxWithTimeout, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to watch TigeraStatus resources: %w", err)
	}
	defer watchInterface.Stop()

	// Set up periodic status check to ensure watch is working
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Wait for all statuses to become available
	for {
		select {
		case <-ctxWithTimeout.Done():
			// Before returning timeout error, do one final check
			if available, err := checkAllTigeraStatusesAvailable(ctx, dynamicClient, tigeraStatusGVR); err == nil && available {
				return nil
			}
			// Get debug info before returning error
			debugInfo := getTigeraStatusDebugInfo(ctx, dynamicClient, tigeraStatusGVR)
			return fmt.Errorf("timeout waiting for Tigera operator status resources to become available: %w. Debug info: %s", ctxWithTimeout.Err(), debugInfo)
		case <-ticker.C:
			// Periodic check to ensure watch is working and provide feedback
			if available, err := checkAllTigeraStatusesAvailable(ctxWithTimeout, dynamicClient, tigeraStatusGVR); err == nil && available {
				return nil
			}
		case event, ok := <-watchInterface.ResultChan():
			if !ok {
				// Channel closed, recreate watch
				watchInterface.Stop()
				var newWatch watch.Interface
				newWatch, err = dynamicClient.Resource(tigeraStatusGVR).Watch(ctxWithTimeout, metav1.ListOptions{})
				if err != nil {
					return fmt.Errorf("failed to recreate watch: %w", err)
				}
				watchInterface = newWatch
				continue
			}

			switch event.Type {
			case watch.Added, watch.Modified:
				// Check if all statuses are available
				if available, err := checkAllTigeraStatusesAvailable(ctxWithTimeout, dynamicClient, tigeraStatusGVR); err != nil {
					return fmt.Errorf("status check failed after event: %w", err)
				} else if available {
					return nil
				}
			case watch.Error:
				return fmt.Errorf("watch error: %v", event.Object)
			}
		}
	}
}

// checkAllTigeraStatusesAvailable checks if all TigeraStatus resources are available
func checkAllTigeraStatusesAvailable(ctx context.Context, dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) (bool, error) {
	// List all TigeraStatus resources (cluster-scoped)
	list, err := dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list TigeraStatus resources: %w", err)
	}

	// If no cluster-scoped resources found, try namespace-scoped resources
	if len(list.Items) == 0 {
		// Try common namespaces where TigeraStatus resources might be located
		namespaces := []string{"tigera-operator", "calico-system", "kube-system"}
		for _, ns := range namespaces {
			nsList, err := dynamicClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
			if err == nil && len(nsList.Items) > 0 {
				list = nsList
				break
			}
		}
	}

	// If still no resources found, consider it as not available
	if len(list.Items) == 0 {
		return false, nil
	}

	// Check each TigeraStatus resource
	for _, item := range list.Items {
		if !isTigeraStatusAvailable(&item) {
			return false, nil
		}
	}

	return true, nil
}

// getTigeraStatusDebugInfo returns debug information about TigeraStatus resources
func getTigeraStatusDebugInfo(ctx context.Context, dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) string {
	// Check cluster-scoped resources first
	list, err := dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Sprintf("Failed to list TigeraStatus resources: %v", err)
	}

	var allItems []unstructured.Unstructured
	if len(list.Items) > 0 {
		allItems = append(allItems, list.Items...)
	}

	// Also check namespace-scoped resources
	namespaces := []string{"tigera-operator", "calico-system", "kube-system"}
	for _, ns := range namespaces {
		nsList, err := dynamicClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
		if err == nil && len(nsList.Items) > 0 {
			allItems = append(allItems, nsList.Items...)
		}
	}

	if len(allItems) == 0 {
		return "No TigeraStatus resources found (cluster-scoped or namespace-scoped)"
	}

	var info []string
	for _, item := range allItems {
		name := item.GetName()
		namespace := item.GetNamespace()
		if namespace == "" {
			namespace = "cluster-scoped"
		}
		status := "Not Available"
		if isTigeraStatusAvailable(&item) {
			status = "Available"
		}
		info = append(info, fmt.Sprintf("%s/%s: %s", namespace, name, status))
	}

	return fmt.Sprintf("Found %d TigeraStatus resources: %s", len(allItems), strings.Join(info, ", "))
}

// isTigeraStatusAvailable checks if a single TigeraStatus resource is available
func isTigeraStatusAvailable(status *unstructured.Unstructured) bool {
	// Get the status field
	statusField, found, err := unstructured.NestedMap(status.Object, "status")
	if err != nil || !found {
		return false
	}

	// Check for available field first (most direct)
	available, found, err := unstructured.NestedBool(statusField, "available")
	if err == nil && found {
		return available
	}

	// Check for state field with "Available" value
	state, found, err := unstructured.NestedString(statusField, "state")
	if err == nil && found {
		return state == "Available"
	}

	// Check for phase field with "Available" value
	phase, found, err := unstructured.NestedString(statusField, "phase")
	if err == nil && found {
		return phase == "Available"
	}

	// Check conditions for Available condition
	conditions, found, err := unstructured.NestedSlice(statusField, "conditions")
	if err == nil && found {
		// Look for Available condition
		for _, cond := range conditions {
			condition, ok := cond.(map[string]any)
			if !ok {
				continue
			}

			condType, found, err := unstructured.NestedString(condition, "type")
			if err != nil || !found {
				continue
			}

			if condType == "Available" {
				condStatus, found, err := unstructured.NestedString(condition, "status")
				if err != nil || !found {
					return false
				}
				return condStatus == "True"
			}
		}
	}

	return false
}

// WaitForTigeraOperatorReadyWithRetry waits with exponential backoff retry
func WaitForTigeraOperatorReadyWithRetry(ctx context.Context, clientset *kubernetes.Clientset, dynamicClient dynamic.Interface, maxRetries int, initialBackoff time.Duration) error {
	backoff := initialBackoff

	for i := 0; i < maxRetries; i++ {
		err := WaitForTigeraOperatorAvailable(ctx, clientset, dynamicClient, 5*time.Minute)
		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		} else {
			return err
		}
	}

	return fmt.Errorf("exceeded maximum retries waiting for Tigera operator")
}
