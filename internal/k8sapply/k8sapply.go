// Package k8sapply provides functionality to apply Kubernetes resources from YAML files, strings, or URLs.
package k8sapply

import (
	"context"
	"fmt"
	"io"
	"log"
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
	apiResourceLists, err := getAPIResources(clientset)
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

func getAPIResources(clientset kubernetes.Interface) ([]*metav1.APIResourceList, error) {
	discoveryClient := clientset.Discovery()
	apiGroups, err := discoveryClient.ServerGroups()
	if err != nil {
		return nil, err
	}
	retval := []*metav1.APIResourceList{}
	for _, group := range apiGroups.Groups {
		for _, version := range group.Versions {
			resources, err := discoveryClient.ServerResourcesForGroupVersion(version.GroupVersion)
			if err != nil {
				// fmt.Printf("Skipping %s: %v\n", version.GroupVersion, err)
				continue
			}
			retval = append(retval, resources)
		}
	}
	return retval, nil
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
	req.Header.Set("User-Agent", "clyde-k8sapply/1.0")

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
	gvk := obj.GroupVersionKind()
	namespace := obj.GetNamespace()

	// Find the API resource for this GVK
	apiResource, err := a.findAPIResourceForGVK(gvk)
	if err != nil {
		return fmt.Errorf("failed to find API resource for GVK %s/%s %s: %w",
			gvk.GroupVersion().String(), gvk.Kind, obj.GetName(), err)
	}

	// Determine if the resource is namespaced
	isNamespaced := apiResource.Namespaced
	if isNamespaced && namespace == "" {
		namespace = "default"
	}

	gvr := gvk.GroupVersion().WithResource(apiResource.Name)
	resourceClient := a.dynamic.Resource(gvr)

	var existing *unstructured.Unstructured
	if isNamespaced {
		// For namespaced resources, use namespace
		existing, err = resourceClient.Namespace(namespace).Get(ctx, obj.GetName(), metav1.GetOptions{})
	} else {
		// For non-namespaced resources, use cluster-scoped client
		existing, err = resourceClient.Get(ctx, obj.GetName(), metav1.GetOptions{})
	}

	if err != nil {
		// Resource doesn't exist, create it
		if isNamespaced {
			_, err = resourceClient.Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create namespaced object %s %s: %w", gvk.Kind, obj.GetName(), err)
			}
		} else {
			_, err = resourceClient.Create(ctx, obj, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create non-namespaced object %s %s: %w", gvk.Kind, obj.GetName(), err)
			}
		}
		a.Logf("Created %s/%s %s", gvk.GroupVersion().String(), gvk.Kind, obj.GetName())
		return nil
	}

	// Resource exists, update it
	obj.SetResourceVersion(existing.GetResourceVersion())
	if isNamespaced {
		_, err = resourceClient.Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	} else {
		_, err = resourceClient.Update(ctx, obj, metav1.UpdateOptions{})
	}
	if err != nil {
		return fmt.Errorf("failed to update %s %s: %w", gvk.Kind, obj.GetName(), err)
	}
	a.Logf("Updated %s/%s %s", gvk.GroupVersion().String(), gvk.Kind, obj.GetName())

	return nil
}

func (a *Applier) findAPIResourceForGVK(gvk schema.GroupVersionKind) (*metav1.APIResource, error) {
	// First pass: exact match
	for _, apiResourceList := range a.resources {
		for _, apiResource := range apiResourceList.APIResources {
			version := apiResource.Version
			if version == "" {
				version = "v1"
			}
			if apiResource.Group == gvk.Group && version == gvk.Version && apiResource.Kind == gvk.Kind {
				return &apiResource, nil
			}
		}
	}

	// Second pass: try to find by kind and group, ignoring version differences
	// This handles cases where the API server might return different versions than expected
	for _, apiResourceList := range a.resources {
		for _, apiResource := range apiResourceList.APIResources {
			if apiResource.Group == gvk.Group && apiResource.Kind == gvk.Kind {
				return &apiResource, nil
			}
		}
	}

	// Third pass: try to find by kind, ignoring group and version differences
	// This handles cases where the API server might return different versions than expected
	for _, apiResourceList := range a.resources {
		for _, apiResource := range apiResourceList.APIResources {
			log.Printf("apiResource: %+v", apiResource)
			if apiResource.Group == "" && apiResource.Version == "" && apiResource.Kind == gvk.Kind {
				return &apiResource, nil
			}
		}
	}

	return nil, fmt.Errorf("api resource not found: GVK: %s | %s | %s", gvk.Group, gvk.Version, gvk.Kind)
}

// buildGVRFromAPIResource constructs a GroupVersionResource from GVK and APIResource
func (a *Applier) buildGVRFromAPIResource(apiResource *metav1.APIResource) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    apiResource.Group,
		Version:  apiResource.Version,
		Resource: apiResource.Name,
	}
}
