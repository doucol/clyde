package calico

import (
	"context"
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
			return false, nil
		}

		// Check conditions
		for _, condition := range crd.Status.Conditions {
			if condition.Type == apiextensionsv1.Established {
				if condition.Status == apiextensionsv1.ConditionTrue {
					return true, nil
				}
			}
		}

		return false, nil
	}

	return wait.PollUntilContextTimeout(ctx, pollInterval, timeout, true, conditionFunc)
}
