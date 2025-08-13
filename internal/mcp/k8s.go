package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesClient wraps the Kubernetes client with additional functionality
type KubernetesClient struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewKubernetesClient creates a new Kubernetes client
func NewKubernetesClient(kubeconfigPath string) (*KubernetesClient, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &KubernetesClient{
		clientset: clientset,
		config:    config,
	}, nil
}

// UpdateKubernetesProvider updates the Kubernetes provider with a real client
func (k *KubernetesProvider) UpdateClient(client *KubernetesClient) {
	k.client = client
}

// ListResources implements ResourceProvider for Kubernetes with real client
func (k *KubernetesProvider) ListResources(ctx context.Context) ([]Resource, error) {
	if k.client == nil {
		return []Resource{}, nil
	}

	client, ok := k.client.(*KubernetesClient)
	if !ok {
		return []Resource{}, nil
	}

	var resources []Resource

	// List pods
	pods, err := client.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, pod := range pods.Items {
			resources = append(resources, Resource{
				URI:         fmt.Sprintf("k8s://pods/%s/%s", pod.Namespace, pod.Name),
				Name:        pod.Name,
				Description: fmt.Sprintf("Pod %s in namespace %s", pod.Name, pod.Namespace),
				MimeType:    "application/json",
				Content:     mustMarshal(pod),
				Metadata: map[string]string{
					"kind":      "Pod",
					"namespace": pod.Namespace,
					"status":    string(pod.Status.Phase),
				},
			})
		}
	}

	// List services
	services, err := client.clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, service := range services.Items {
			resources = append(resources, Resource{
				URI:         fmt.Sprintf("k8s://services/%s/%s", service.Namespace, service.Name),
				Name:        service.Name,
				Description: fmt.Sprintf("Service %s in namespace %s", service.Name, service.Namespace),
				MimeType:    "application/json",
				Content:     mustMarshal(service),
				Metadata: map[string]string{
					"kind":      "Service",
					"namespace": service.Namespace,
					"type":      string(service.Spec.Type),
				},
			})
		}
	}

	// List deployments
	deployments, err := client.clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, deployment := range deployments.Items {
			resources = append(resources, Resource{
				URI:         fmt.Sprintf("k8s://deployments/%s/%s", deployment.Namespace, deployment.Name),
				Name:        deployment.Name,
				Description: fmt.Sprintf("Deployment %s in namespace %s", deployment.Name, deployment.Namespace),
				MimeType:    "application/json",
				Content:     mustMarshal(deployment),
				Metadata: map[string]string{
					"kind":      "Deployment",
					"namespace": deployment.Namespace,
					"replicas":  fmt.Sprintf("%d", *deployment.Spec.Replicas),
				},
			})
		}
	}

	// List nodes
	nodes, err := client.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err == nil {
		for _, node := range nodes.Items {
			resources = append(resources, Resource{
				URI:         fmt.Sprintf("k8s://nodes/%s", node.Name),
				Name:        node.Name,
				Description: fmt.Sprintf("Node %s", node.Name),
				MimeType:    "application/json",
				Content:     mustMarshal(node),
				Metadata: map[string]string{
					"kind":   "Node",
					"status": string(node.Status.Conditions[len(node.Status.Conditions)-1].Type),
				},
			})
		}
	}

	return resources, nil
}

// GetResource implements ResourceProvider for Kubernetes with real client
func (k *KubernetesProvider) GetResource(ctx context.Context, uri string) (*Resource, error) {
	if k.client == nil {
		return nil, ErrResourceNotFound
	}

	client, ok := k.client.(*KubernetesClient)
	if !ok {
		return nil, ErrResourceNotFound
	}

	// Parse URI: k8s://kind/namespace/name or k8s://kind/name
	parts := strings.Split(strings.TrimPrefix(uri, "k8s://"), "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid Kubernetes resource URI: %s", uri)
	}

	kind := parts[0]
	var namespace, name string

	if len(parts) == 3 {
		namespace = parts[1]
		name = parts[2]
	} else if len(parts) == 2 {
		name = parts[1]
	}

	switch kind {
	case "pods":
		if namespace == "" {
			return nil, fmt.Errorf("namespace required for pods")
		}
		pod, err := client.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return &Resource{
			URI:         uri,
			Name:        pod.Name,
			Description: fmt.Sprintf("Pod %s in namespace %s", pod.Name, pod.Namespace),
			MimeType:    "application/json",
			Content:     mustMarshal(pod),
			Metadata: map[string]string{
				"kind":      "Pod",
				"namespace": pod.Namespace,
				"status":    string(pod.Status.Phase),
			},
		}, nil

	case "services":
		if namespace == "" {
			return nil, fmt.Errorf("namespace required for services")
		}
		service, err := client.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return &Resource{
			URI:         uri,
			Name:        service.Name,
			Description: fmt.Sprintf("Service %s in namespace %s", service.Name, service.Namespace),
			MimeType:    "application/json",
			Content:     mustMarshal(service),
			Metadata: map[string]string{
				"kind":      "Service",
				"namespace": service.Namespace,
				"type":      string(service.Spec.Type),
			},
		}, nil

	case "deployments":
		if namespace == "" {
			return nil, fmt.Errorf("namespace required for deployments")
		}
		deployment, err := client.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return &Resource{
			URI:         uri,
			Name:        deployment.Name,
			Description: fmt.Sprintf("Deployment %s in namespace %s", deployment.Name, deployment.Namespace),
			MimeType:    "application/json",
			Content:     mustMarshal(deployment),
			Metadata: map[string]string{
				"kind":      "Deployment",
				"namespace": deployment.Namespace,
				"replicas":  fmt.Sprintf("%d", *deployment.Spec.Replicas),
			},
		}, nil

	case "nodes":
		node, err := client.clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return &Resource{
			URI:         uri,
			Name:        node.Name,
			Description: fmt.Sprintf("Node %s", node.Name),
			MimeType:    "application/json",
			Content:     mustMarshal(node),
			Metadata: map[string]string{
				"kind":   "Node",
				"status": string(node.Status.Conditions[len(node.Status.Conditions)-1].Type),
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported Kubernetes resource kind: %s", kind)
	}
}

// UpdateKubernetesTools updates the Kubernetes tools with a real client
func (k *KubernetesTools) UpdateClient(client *KubernetesClient) {
	k.client = client
}

// CallTool implements ToolProvider for Kubernetes with real client
func (k *KubernetesTools) CallTool(ctx context.Context, name string, arguments map[string]any) (any, error) {
	if k.client == nil {
		return nil, ErrToolNotFound
	}

	client, ok := k.client.(*KubernetesClient)
	if !ok {
		return nil, ErrToolNotFound
	}

	switch name {
	case "get_pod_logs":
		return k.getPodLogs(ctx, client, arguments)
	case "describe_pod":
		return k.describePod(ctx, client, arguments)
	case "get_node_status":
		return k.getNodeStatus(ctx, client, arguments)
	default:
		return nil, ErrToolNotFound
	}
}

// getPodLogs retrieves logs from a Kubernetes pod
func (k *KubernetesTools) getPodLogs(ctx context.Context, client *KubernetesClient, arguments map[string]any) (any, error) {
	namespace, ok := arguments["namespace"].(string)
	if !ok {
		return nil, ErrInvalidToolArguments
	}

	podName, ok := arguments["pod"].(string)
	if !ok {
		return nil, ErrInvalidToolArguments
	}

	containerName, _ := arguments["container"].(string)
	tailLines := int64(100)
	if tl, ok := arguments["tailLines"].(float64); ok {
		tailLines = int64(tl)
	}

	req := client.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &tailLines,
	})

	logs, err := req.Do(ctx).Raw()
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}

	return map[string]interface{}{
		"logs":      string(logs),
		"pod":       podName,
		"namespace": namespace,
		"container": containerName,
		"tailLines": tailLines,
	}, nil
}

// describePod gets detailed information about a Kubernetes pod
func (k *KubernetesTools) describePod(ctx context.Context, client *KubernetesClient, arguments map[string]any) (any, error) {
	namespace, ok := arguments["namespace"].(string)
	if !ok {
		return nil, ErrInvalidToolArguments
	}

	podName, ok := arguments["pod"].(string)
	if !ok {
		return nil, ErrInvalidToolArguments
	}

	pod, err := client.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	return map[string]interface{}{
		"pod":       pod,
		"namespace": namespace,
		"status":    pod.Status,
		"spec":      pod.Spec,
	}, nil
}

// getNodeStatus gets the status of Kubernetes nodes
func (k *KubernetesTools) getNodeStatus(ctx context.Context, client *KubernetesClient, arguments map[string]any) (any, error) {
	var nodeName string
	if node, ok := arguments["node"].(string); ok {
		nodeName = node
	}

	if nodeName != "" {
		// Get specific node
		node, err := client.clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get node: %w", err)
		}
		return map[string]interface{}{
			"node":  node,
			"status": node.Status,
		}, nil
	}

	// Get all nodes
	nodes, err := client.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var nodeStatuses []map[string]interface{}
	for _, node := range nodes.Items {
		nodeStatuses = append(nodeStatuses, map[string]interface{}{
			"name":   node.Name,
			"status": node.Status,
		})
	}

	return map[string]interface{}{
		"nodes": nodeStatuses,
		"count": len(nodeStatuses),
	}, nil
}

// mustMarshal safely marshals an object to JSON
func mustMarshal(obj interface{}) []byte {
	data, err := json.Marshal(obj)
	if err != nil {
		return []byte(fmt.Sprintf(`{"error": "failed to marshal: %v"}`, err))
	}
	return data
}
