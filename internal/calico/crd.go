package calico

import (
	"context"
	"fmt"
	"strings"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func WaitForCRDEstablished(ctx context.Context, apiextClient apiextensionsclient.Interface, crdName string, timeout time.Duration) error {
	pollInterval := 2 * time.Second

	conditionFunc := func(ctx context.Context) (bool, error) {
		crd, err := apiextClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
		if err != nil {
			// If the CRD doesn't exist yet, return false to continue waiting
			return false, nil
		}

		// Check if the CRD has status conditions
		if crd.Status.Conditions == nil {
			return false, nil
		}

		// Check for Established condition
		for _, condition := range crd.Status.Conditions {
			if condition.Type == apiextensionsv1.Established {
				if condition.Status == apiextensionsv1.ConditionTrue {
					return true, nil
				}
				// If the condition exists but is not True, continue waiting
				return false, nil
			}
		}

		// If no Established condition found, the CRD is not ready yet
		return false, nil
	}

	// Use exponential backoff with jitter for more robust waiting
	return wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, conditionFunc)
}

// WaitForCRDReady waits for a CRD to be both established and ready for use
func WaitForCRDReady(ctx context.Context, apiextClient apiextensionsclient.Interface, crdName string, timeout time.Duration) error {
	// First wait for the CRD to be established
	if err := WaitForCRDEstablished(ctx, apiextClient, crdName, timeout); err != nil {
		return fmt.Errorf("CRD %s did not become established: %w", crdName, err)
	}

	// Additional wait to ensure the CRD is fully ready
	// This includes waiting for the API server to fully process the CRD
	readyTimeout := 30 * time.Second
	if timeout < readyTimeout {
		readyTimeout = timeout / 2
	}

	readyConditionFunc := func(ctx context.Context) (bool, error) {
		crd, err := apiextClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		// Check if the CRD has status conditions
		if crd.Status.Conditions == nil {
			return false, nil
		}

		// Check for both Established and NamesAccepted conditions
		established := false
		namesAccepted := false

		for _, condition := range crd.Status.Conditions {
			switch condition.Type {
			case apiextensionsv1.Established:
				if condition.Status == apiextensionsv1.ConditionTrue {
					established = true
				}
			case apiextensionsv1.NamesAccepted:
				if condition.Status == apiextensionsv1.ConditionTrue {
					namesAccepted = true
				}
			}
		}

		// Both conditions must be true for the CRD to be fully ready
		return established && namesAccepted, nil
	}

	return wait.PollUntilContextTimeout(ctx, 1*time.Second, readyTimeout, true, readyConditionFunc)
}

// WaitForMultipleCRDsReady waits for multiple CRDs to be ready
func WaitForMultipleCRDsReady(ctx context.Context, apiextClient apiextensionsclient.Interface, crdNames []string, timeout time.Duration) error {
	// Calculate timeout per CRD
	timeoutPerCRD := timeout / time.Duration(len(crdNames))
	if timeoutPerCRD < 30*time.Second {
		timeoutPerCRD = 30 * time.Second
	}

	for _, crdName := range crdNames {
		if err := WaitForCRDReady(ctx, apiextClient, crdName, timeoutPerCRD); err != nil {
			return fmt.Errorf("CRD %s did not become ready: %w", crdName, err)
		}
	}

	return nil
}

// GetCRDStatus returns detailed status information for a specific CRD
func GetCRDStatus(ctx context.Context, apiextClient apiextensionsclient.Interface, crdName string) (*CRDStatus, error) {
	crd, err := apiextClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get CRD %s: %w", crdName, err)
	}

	status := &CRDStatus{
		Name:               crdName,
		Established:        false,
		NamesAccepted:      false,
		Conditions:         make(map[string]string),
		LastTransitionTime: nil,
	}

	// Check all conditions
	for _, condition := range crd.Status.Conditions {
		status.Conditions[string(condition.Type)] = string(condition.Status)

		switch condition.Type {
		case apiextensionsv1.Established:
			status.Established = condition.Status == apiextensionsv1.ConditionTrue
			status.LastTransitionTime = &condition.LastTransitionTime
		case apiextensionsv1.NamesAccepted:
			status.NamesAccepted = condition.Status == apiextensionsv1.ConditionTrue
		}
	}

	return status, nil
}

// CRDStatus represents the status of a CustomResourceDefinition
type CRDStatus struct {
	Name               string
	Established        bool
	NamesAccepted      bool
	Conditions         map[string]string
	LastTransitionTime *metav1.Time
}

// IsReady returns true if the CRD is fully ready
func (cs *CRDStatus) IsReady() bool {
	return cs.Established && cs.NamesAccepted
}

// GetStatusSummary returns a human-readable summary of the CRD status
func (cs *CRDStatus) GetStatusSummary() string {
	if cs.IsReady() {
		return "Ready"
	}

	var issues []string
	if !cs.Established {
		issues = append(issues, "Not Established")
	}
	if !cs.NamesAccepted {
		issues = append(issues, "Names Not Accepted")
	}

	return fmt.Sprintf("Not Ready: %s", strings.Join(issues, ", "))
}
