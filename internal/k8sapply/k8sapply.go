// Package k8sapply provides functionality to apply Kubernetes resources from various sources
package k8sapply

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
)

// Applier handles applying Kubernetes resources
type Applier struct {
	dynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	mapper          meta.RESTMapper
	logWriter       io.Writer
	logChan         chan string
}

// ApplyOptions configures how resources are applied
type ApplyOptions struct {
	// Namespace to apply resources to (overrides resource namespace if set)
	Namespace string
	// DryRun performs a dry run without actually applying resources
	DryRun bool
	// Force will delete and recreate resources if update fails
	Force bool
	// Timeout for HTTP requests when fetching from URL
	HTTPTimeout time.Duration
}

// ApplyResult contains information about an applied resource
type ApplyResult struct {
	Name      string
	Namespace string
	Kind      string
	Action    string // "created", "updated", "unchanged", "deleted"
	Error     error
}

// NewApplierWithClients creates a new Applier with existing Kubernetes clients
func NewApplierWithClients(clientset kubernetes.Interface, dynamicClient dynamic.Interface, logger io.Writer) (*Applier, error) {
	// Use the clientset's discovery client
	discoveryClient := clientset.Discovery()

	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get API group resources: %w", err)
	}

	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	return &Applier{
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		mapper:          mapper,
		logWriter:       logger,
	}, nil
}

func (a *Applier) LogChan() chan string {
	if a.logChan == nil {
		a.logChan = make(chan string, 100) // Buffered channel to avoid blocking
	}
	return a.logChan
}

func (a *Applier) Logf(format string, args ...any) {
	if len(format) == 0 {
		a.Log(fmt.Sprintf(format, args...))
	}
}

func (a *Applier) Log(message string) {
	if len(message) == 0 {
		return
	}
	if a.logWriter != nil {
		if _, err := fmt.Fprint(a.logWriter, message+"\n"); err != nil {
			panic(fmt.Sprintf("failed to write log: %v", err))
		}
	}
	if a.logChan != nil {
		a.logChan <- message
	}
}

func (a *Applier) Close() {
	if a.logChan != nil {
		close(a.logChan)
	}
}

// ApplyFromURL applies resources from a YAML URL
func (a *Applier) ApplyFromURL(ctx context.Context, url string, opts *ApplyOptions) ([]ApplyResult, error) {
	if opts == nil {
		opts = &ApplyOptions{}
	}

	yamlContent, err := a.fetchFromURL(url, opts.HTTPTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from URL: %w", err)
	}

	return a.ApplyFromString(ctx, yamlContent, opts)
}

// ApplyFromFile applies resources from a YAML file
func (a *Applier) ApplyFromFile(ctx context.Context, filepath string, opts *ApplyOptions) ([]ApplyResult, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return a.ApplyFromString(ctx, string(content), opts)
}

// ApplyFromString applies resources from a YAML string
func (a *Applier) ApplyFromString(ctx context.Context, yamlContent string, opts *ApplyOptions) ([]ApplyResult, error) {
	if opts == nil {
		opts = &ApplyOptions{}
	}

	resources, err := a.parseYAML(yamlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	results := make([]ApplyResult, 0, len(resources))
	for _, resource := range resources {
		result := a.applyResource(ctx, resource, opts)
		if result.Error != nil {
			a.Logf("Error applying %s/%s (%s): %v", resource.GetNamespace(), resource.GetName(), resource.GetKind(), result.Error)
		} else {
			a.Logf("Applied %s/%s (%s): %s", resource.GetNamespace(), resource.GetName(), resource.GetKind(), result.Action)
		}
		results = append(results, result)
	}

	return results, nil
}

// fetchFromURL fetches YAML content from a URL
func (a *Applier) fetchFromURL(url string, timeout time.Duration) (string, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// parseYAML parses YAML content into unstructured objects
func (a *Applier) parseYAML(yamlContent string) ([]*unstructured.Unstructured, error) {
	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(yamlContent)), 4096)
	resources := make([]*unstructured.Unstructured, 0)

	for {
		var rawObj runtime.RawExtension
		err := decoder.Decode(&rawObj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error decoding YAML: %w", err)
		}

		if rawObj.Raw == nil {
			continue
		}

		obj, _, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).Decode(rawObj.Raw, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("error decoding raw object: %w", err)
		}

		unstructuredObj, ok := obj.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("object is not unstructured")
		}

		// Skip empty objects
		if unstructuredObj.GetKind() == "" {
			continue
		}

		resources = append(resources, unstructuredObj)
	}

	return resources, nil
}

// applyResource applies a single resource
func (a *Applier) applyResource(ctx context.Context, resource *unstructured.Unstructured, opts *ApplyOptions) ApplyResult {
	result := ApplyResult{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
		Kind:      resource.GetKind(),
	}

	// Override namespace if specified in options
	if opts.Namespace != "" && resource.GetNamespace() != "" {
		resource.SetNamespace(opts.Namespace)
		result.Namespace = opts.Namespace
	}

	// Get GVR for the resource
	gvk := resource.GroupVersionKind()
	mapping, err := a.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		result.Error = fmt.Errorf("failed to get REST mapping: %w", err)
		return result
	}

	// Get the appropriate client
	var dr dynamic.ResourceInterface
	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if resource.GetNamespace() == "" {
			resource.SetNamespace("default")
			result.Namespace = "default"
		}
		dr = a.dynamicClient.Resource(mapping.Resource).Namespace(resource.GetNamespace())
	} else {
		dr = a.dynamicClient.Resource(mapping.Resource)
	}

	// Try to get existing resource
	existing, err := dr.Get(ctx, resource.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create new resource
			if opts.DryRun {
				result.Action = "would create"
				return result
			}

			_, err = dr.Create(ctx, resource, metav1.CreateOptions{})
			if err != nil {
				result.Error = fmt.Errorf("failed to create resource: %w", err)
				return result
			}
			result.Action = "created"
			return result
		}
		result.Error = fmt.Errorf("failed to get resource: %w", err)
		return result
	}

	// Update existing resource
	resource.SetResourceVersion(existing.GetResourceVersion())

	if opts.DryRun {
		result.Action = "would update"
		return result
	}

	_, err = dr.Update(ctx, resource, metav1.UpdateOptions{})
	if err != nil {
		if opts.Force {
			// Force update by deleting and recreating
			err = dr.Delete(ctx, resource.GetName(), metav1.DeleteOptions{})
			if err != nil {
				result.Error = fmt.Errorf("failed to delete for force update: %w", err)
				return result
			}

			// Remove resource version for create
			resource.SetResourceVersion("")
			_, err = dr.Create(ctx, resource, metav1.CreateOptions{})
			if err != nil {
				result.Error = fmt.Errorf("failed to recreate resource: %w", err)
				return result
			}
			result.Action = "recreated"
			return result
		}
		result.Error = fmt.Errorf("failed to update resource: %w", err)
		return result
	}

	result.Action = "updated"
	return result
}

// ResultsSummary provides a summary of apply results
func ResultsSummary(results []ApplyResult) string {
	var created, updated, unchanged, failed int
	var errors []string

	for _, r := range results {
		if r.Error != nil {
			failed++
			errors = append(errors, fmt.Sprintf("%s/%s (%s): %v", r.Namespace, r.Name, r.Kind, r.Error))
			continue
		}

		switch r.Action {
		case "created", "would create":
			created++
		case "updated", "would update", "recreated":
			updated++
		case "unchanged":
			unchanged++
		}
	}

	summary := fmt.Sprintf("Created: %d, Updated: %d, Unchanged: %d, Failed: %d",
		created, updated, unchanged, failed)

	if len(errors) > 0 {
		summary += "\nErrors:\n" + strings.Join(errors, "\n")
	}

	return summary
}
