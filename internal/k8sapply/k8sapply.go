// Package k8sapply provides functionality to apply Kubernetes resources from YAML files, strings, or URLs.
package k8sapply

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Applier handles applying Kubernetes resources from YAML
type Applier struct {
	clientset kubernetes.Interface
	dynamic   dynamic.Interface
	logger    io.Writer
	resources []*metav1.APIResourceList
}

// NewApplier creates a new Applier instance
func NewApplier(clientset kubernetes.Interface, dynamicClient dynamic.Interface, logger io.Writer) (*Applier, error) {
	// Get all API resources
	_, apiResourceLists, err := clientset.Discovery().ServerGroupsAndResources()
	if err != nil {
		return nil, err
	}

	return &Applier{
		clientset: clientset,
		dynamic:   dynamicClient,
		logger:    logger,
		resources: apiResourceLists,
	}, nil
}

func (a *Applier) Logf(format string, args ...any) {
	if a.logger != nil {
		_, _ = fmt.Fprintf(a.logger, format+"\n", args...)
	}
}

// ApplyFile applies resources from a YAML file
func (a *Applier) ApplyFile(ctx context.Context, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	return a.ApplyReader(ctx, file)
}

// ApplyString applies resources from a YAML string
func (a *Applier) ApplyString(ctx context.Context, yamlContent string) error {
	reader := strings.NewReader(yamlContent)
	return a.ApplyReader(ctx, reader)
}

// ApplyURL applies resources from a YAML file located at a URL
func (a *Applier) ApplyURL(ctx context.Context, url string) error {
	return a.ApplyURLWithTimeout(ctx, url, 30*time.Second)
}

// ApplyURLWithTimeout applies resources from a YAML file located at a URL with a custom timeout
func (a *Applier) ApplyURLWithTimeout(ctx context.Context, url string, timeout time.Duration) error {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request for %s: %w", url, err)
	}

	// Set User-Agent to identify the client
	req.Header.Set("User-Agent", "k8sapply/1.0")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch YAML from %s: %w", url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP request failed with status %d for %s", resp.StatusCode, url)
	}

	// Apply the YAML content from the response body
	return a.ApplyReader(ctx, resp.Body)
}

// ApplyReader applies resources from an io.Reader containing YAML
func (a *Applier) ApplyReader(ctx context.Context, reader io.Reader) error {
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 1024)

	for {
		var obj unstructured.Unstructured
		err := decoder.Decode(&obj)
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode YAML: %w", err)
		}

		// Skip empty objects
		if len(obj.Object) == 0 {
			continue
		}

		// Apply the resource
		if err := a.applyResource(ctx, &obj); err != nil {
			return fmt.Errorf("failed to apply resource %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}
	}

	return nil
}

// applyResource applies a single Kubernetes resource
func (a *Applier) applyResource(ctx context.Context, obj *unstructured.Unstructured) error {
	gvk := obj.GetObjectKind().GroupVersionKind()
	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	var err error
	var gvr schema.GroupVersionResource
	if gvr, err = a.findGVRForGVK(gvk); err != nil {
		return err
	}

	// Get the dynamic client for this resource
	resourceClient := a.dynamic.Resource(gvr)

	// Check if resource exists
	existing, err := resourceClient.Namespace(namespace).Get(ctx, obj.GetName(), metav1.GetOptions{})
	if err != nil {
		// Resource doesn't exist, create it
		_, err = resourceClient.Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create %s %s: %w", gvk.Kind, obj.GetName(), err)
		}
		a.Logf("Created %s/%s %s", gvk.GroupVersion().String(), gvk.Kind, obj.GetName())
		return nil
	}

	// Resource exists, update it
	obj.SetResourceVersion(existing.GetResourceVersion())
	_, err = resourceClient.Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update %s %s: %w", gvk.Kind, obj.GetName(), err)
	}
	a.Logf("Updated %s/%s %s", gvk.GroupVersion().String(), gvk.Kind, obj.GetName())

	return nil
}

func (a *Applier) findGVRForGVK(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	for _, apiResourceList := range a.resources {
		for _, apiResource := range apiResourceList.APIResources {
			if apiResource.Group == gvk.Group && apiResource.Version == gvk.Version && apiResource.Kind == gvk.Kind {
				return schema.GroupVersionResource{
					Group:    apiResource.Group,
					Version:  apiResource.Version,
					Resource: apiResource.Name,
				}, nil
			}
		}
	}

	return schema.GroupVersionResource{}, fmt.Errorf("kind %s not found", gvk.Kind)
}
