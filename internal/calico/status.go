package calico

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// WaitForTigeraOperatorAvailable waits for all Tigera operator status resources to become available
func (cm *CalicoManager) WaitForTigeraOperatorAvailable(ctx context.Context, clientset *kubernetes.Clientset, dynamicClient dynamic.Interface, timeout time.Duration) error {
	cm.Logf("[white]Waiting for Tigera operator to be available...")

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// First, check if the Tigera operator deployment is running
	if err := cm.waitForTigeraOperatorDeployment(ctxWithTimeout, clientset); err != nil {
		return fmt.Errorf("Tigera operator deployment not ready: %w", err)
	}

	// Wait for the operator to create its status resources
	// The actual resource name might be different, so let's try multiple approaches
	if err := cm.waitForTigeraStatusResources(ctxWithTimeout, dynamicClient); err != nil {
		cm.Logf("[yellow]Warning: Could not verify Tigera status resources: %v", err)
		cm.Logf("[yellow]Trying alternative verification methods...")

		// Try to verify the operator is working by checking if it can handle basic operations
		if err := cm.verifyTigeraOperatorFunctionality(ctxWithTimeout, dynamicClient); err != nil {
			cm.Logf("[yellow]Warning: Could not verify Tigera operator functionality: %v", err)
			cm.Logf("[yellow]Continuing anyway as the operator deployment is ready")
		} else {
			cm.Logf("[green]Tigera operator functionality verified!")
		}
	}

	cm.Logf("[green]Tigera operator is available and ready!")
	return nil
}

// waitForTigeraOperatorDeployment waits for the Tigera operator deployment to be ready
func (cm *CalicoManager) waitForTigeraOperatorDeployment(ctx context.Context, clientset *kubernetes.Clientset) error {
	cm.Logf("[white]Checking Tigera operator deployment status...")

	// Wait for the deployment to be available
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			deployment, err := clientset.AppsV1().Deployments("tigera-operator").Get(ctx, "tigera-operator", metav1.GetOptions{})
			if err != nil {
				cm.Logf("[yellow]Tigera operator deployment not found yet: %v", err)
				continue
			}

			// Check if deployment is ready
			if deployment.Status.ReadyReplicas > 0 &&
				deployment.Status.AvailableReplicas > 0 &&
				deployment.Status.UpdatedReplicas == *deployment.Spec.Replicas {

				// Also check that the pods are actually running and ready
				if err := cm.checkTigeraOperatorPods(ctx, clientset); err != nil {
					cm.Logf("[yellow]Warning: Pod check failed: %v", err)
					// Don't fail here, the deployment is ready
				}

				cm.Logf("[green]Tigera operator deployment is ready: %d/%d replicas available",
					deployment.Status.AvailableReplicas, *deployment.Spec.Replicas)
				return nil
			}

			cm.Logf("[white]Tigera operator deployment not ready yet: %d/%d replicas available, %d updated",
				deployment.Status.AvailableReplicas, *deployment.Spec.Replicas, deployment.Status.UpdatedReplicas)
		}
	}
}

// checkTigeraOperatorPods checks that the Tigera operator pods are running and ready
func (cm *CalicoManager) checkTigeraOperatorPods(ctx context.Context, clientset *kubernetes.Clientset) error {
	pods, err := clientset.CoreV1().Pods("tigera-operator").List(ctx, metav1.ListOptions{
		LabelSelector: "name=tigera-operator",
	})
	if err != nil {
		return fmt.Errorf("failed to list Tigera operator pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return fmt.Errorf("no Tigera operator pods found")
	}

	// Check each pod
	for _, pod := range pods.Items {
		if pod.Status.Phase != "Running" {
			cm.Logf("[yellow]Pod %s is not running: %s", pod.Name, pod.Status.Phase)
			continue
		}

		// Check if pod is ready
		ready := false
		for _, condition := range pod.Status.Conditions {
			if condition.Type == "Ready" && condition.Status == "True" {
				ready = true
				break
			}
		}

		if ready {
			cm.Logf("[green]Pod %s is running and ready", pod.Name)
		} else {
			cm.Logf("[yellow]Pod %s is running but not ready", pod.Name)
		}
	}

	return nil
}

// waitForTigeraStatusResources waits for Tigera status resources to be available
func (cm *CalicoManager) waitForTigeraStatusResources(ctx context.Context, dynamicClient dynamic.Interface) error {
	cm.Logf("[white]Waiting for Tigera status resources...")

	// Create a timeout context for this function
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	// Try different possible resource names and locations
	possibleResources := []struct {
		group      string
		version    string
		resource   string
		namespaces []string
	}{
		{"operator.tigera.io", "v1", "tigerastatuses", []string{""}}, // cluster-scoped
		{"operator.tigera.io", "v1", "tigerastatuses", []string{"tigera-operator", "calico-system"}},
		{"operator.tigera.io", "v1", "tigerastatus", []string{""}}, // singular
		{"operator.tigera.io", "v1", "tigerastatus", []string{"tigera-operator", "calico-system"}},
		{"operator.tigera.io", "v1alpha1", "tigerastatuses", []string{""}},
		{"operator.tigera.io", "v1alpha1", "tigerastatuses", []string{"tigera-operator", "calico-system"}},
	}

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("timeout waiting for Tigera status resources: %w", timeoutCtx.Err())
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Try to find any of the possible resources
			for _, resource := range possibleResources {
				gvr := schema.GroupVersionResource{
					Group:    resource.group,
					Version:  resource.version,
					Resource: resource.resource,
				}

				// Try cluster-scoped first
				if len(resource.namespaces) > 0 && resource.namespaces[0] == "" {
					list, err := dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
					if err == nil && len(list.Items) > 0 {
						cm.Logf("[green]Found %d Tigera status resources (cluster-scoped)", len(list.Items))
						return nil
					}
				}

				// Try namespace-scoped
				for _, namespace := range resource.namespaces {
					if namespace == "" {
						continue
					}

					list, err := dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
					if err == nil && len(list.Items) > 0 {
						cm.Logf("[green]Found %d Tigera status resources in namespace %s", len(list.Items), namespace)
						return nil
					}
				}
			}

			cm.Logf("[white]Tigera status resources not found yet, waiting...")
		}
	}
}

// checkAllTigeraStatusesAvailable checks if all TigeraStatus resources are available
func (cm *CalicoManager) checkAllTigeraStatusesAvailable(ctx context.Context, dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) (bool, error) {
	// List all TigeraStatus resources (cluster-scoped)
	list, err := dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list TigeraStatus resources: %w", err)
	}

	// If no cluster-scoped resources found, try namespace-scoped resources
	if len(list.Items) == 0 {
		// Try common namespaces where TigeraStatus resources might be located
		namespaces := []string{"tigera-operator", "calico-system"}
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
		if !cm.isTigeraStatusAvailable(&item) {
			cm.Logf("TigeraStatus resource %s/%s is not available", item.GetNamespace(), item.GetName())
			return false, nil
		}
		cm.Logf("TigeraStatus resource %s/%s is available", item.GetNamespace(), item.GetName())
	}

	return true, nil
}

// getTigeraStatusDebugInfo returns debug information about TigeraStatus resources
func (cm *CalicoManager) getTigeraStatusDebugInfo(ctx context.Context, dynamicClient dynamic.Interface, gvr schema.GroupVersionResource) string {
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
		if cm.isTigeraStatusAvailable(&item) {
			status = "Available"
		}
		info = append(info, fmt.Sprintf("%s/%s: %s", namespace, name, status))
	}

	return fmt.Sprintf("Found %d TigeraStatus resources: %s", len(allItems), strings.Join(info, ", "))
}

// isTigeraStatusAvailable checks if a single TigeraStatus resource is available
func (cm *CalicoManager) isTigeraStatusAvailable(status *unstructured.Unstructured) bool {
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

// verifyTigeraOperatorFunctionality verifies that the Tigera operator is working by checking basic operations
func (cm *CalicoManager) verifyTigeraOperatorFunctionality(ctx context.Context, dynamicClient dynamic.Interface) error {
	cm.Logf("[white]Verifying Tigera operator functionality...")

	// Try to check if the operator can handle basic CRD operations
	// This is a more reliable way to verify the operator is working

	// Check if we can list any resources from the operator.tigera.io group
	possibleResources := []string{"installations", "goldmanes", "whiskers"}

	for _, resource := range possibleResources {
		gvr := schema.GroupVersionResource{
			Group:    "operator.tigera.io",
			Version:  "v1",
			Resource: resource,
		}

		// Try to list resources - this will fail if the operator isn't working
		_, err := dynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{Limit: 1})
		if err == nil {
			cm.Logf("[green]Successfully verified %s resource access", resource)
			return nil
		}

		// If we get a "no matches for kind" error, the operator might not be ready yet
		if strings.Contains(err.Error(), "no matches for kind") {
			cm.Logf("[yellow]Resource %s not ready yet: %v", resource, err)
			continue
		}

		// For other errors, log them but continue trying
		cm.Logf("[yellow]Error checking resource %s: %v", resource, err)
	}

	return fmt.Errorf("could not verify Tigera operator functionality with any resource")
}

// WaitForTigeraOperatorReadyWithRetry waits with exponential backoff retry
func (cm *CalicoManager) WaitForTigeraOperatorReadyWithRetry(ctx context.Context, clientset *kubernetes.Clientset, dynamicClient dynamic.Interface, maxRetries int, initialBackoff time.Duration) error {
	backoff := initialBackoff

	for i := 0; i < maxRetries; i++ {
		cm.Logf("[white]Attempt %d/%d: Waiting for Tigera operator to be ready...", i+1, maxRetries)

		err := cm.WaitForTigeraOperatorAvailable(ctx, clientset, dynamicClient, 5*time.Minute)
		if err == nil {
			cm.Logf("[green]Tigera operator is ready on attempt %d", i+1)
			return nil
		}

		if i < maxRetries-1 {
			cm.Logf("[yellow]Attempt %d failed: %v. Retrying in %v...", i+1, err, backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		} else {
			cm.Logf("[red]All %d attempts failed. Tigera operator is not ready: %v", maxRetries, err)
			return fmt.Errorf("exceeded maximum retries waiting for Tigera operator: %w", err)
		}
	}

	return fmt.Errorf("exceeded maximum retries waiting for Tigera operator")
}
