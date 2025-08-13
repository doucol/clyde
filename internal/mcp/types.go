package mcp

import (
	"context"
)

// Resource represents a Kubernetes or Calico resource
type Resource struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	MimeType    string            `json:"mimeType"`
	Content     []byte            `json:"content"`
	Metadata    map[string]string `json:"metadata"`
}

// Tool represents an available tool for the MCP server
type Tool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputSchema map[string]any    `json:"inputSchema"`
	Metadata    map[string]string `json:"metadata"`
}

// ResourceProvider defines the interface for providing resources
type ResourceProvider interface {
	ListResources(ctx context.Context) ([]Resource, error)
	GetResource(ctx context.Context, uri string) (*Resource, error)
}

// ToolProvider defines the interface for providing tools
type ToolProvider interface {
	ListTools(ctx context.Context) ([]Tool, error)
	CallTool(ctx context.Context, name string, arguments map[string]any) (any, error)
}

// KubernetesProvider provides Kubernetes resources
type KubernetesProvider struct {
	client interface{} // Will be k8s.io/client-go/kubernetes.Interface
}

// CalicoProvider provides Calico resources
type CalicoProvider struct {
	calicoClient interface{} // Will be the Calico client from internal/calico
}

// NewKubernetesProvider creates a new Kubernetes resource provider
func NewKubernetesProvider(client interface{}) *KubernetesProvider {
	return &KubernetesProvider{
		client: client,
	}
}

// NewCalicoProvider creates a new Calico resource provider
func NewCalicoProvider(calicoClient interface{}) *CalicoProvider {
	return &CalicoProvider{
		calicoClient: calicoClient,
	}
}



// KubernetesTools provides Kubernetes-related tools
type KubernetesTools struct {
	client interface{} // Will be k8s.io/client-go/kubernetes.Interface
}

// CalicoTools provides Calico-related tools
type CalicoTools struct {
	calicoClient interface{} // Will be the Calico client from internal/calico
}

// NewKubernetesTools creates a new Kubernetes tools provider
func NewKubernetesTools(client interface{}) *KubernetesTools {
	return &KubernetesTools{
		client: client,
	}
}

// NewCalicoTools creates a new Calico tools provider
func NewCalicoTools(calicoClient interface{}) *CalicoTools {
	return &CalicoTools{
		calicoClient: calicoClient,
	}
}

// ListTools implements ToolProvider for Kubernetes
func (k *KubernetesTools) ListTools(ctx context.Context) ([]Tool, error) {
	return []Tool{
		{
			Name:        "get_pod_logs",
			Description: "Get logs from a Kubernetes pod",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"namespace": map[string]any{"type": "string"},
					"pod":       map[string]any{"type": "string"},
					"container": map[string]any{"type": "string"},
					"tailLines": map[string]any{"type": "integer"},
				},
				"required": []string{"namespace", "pod"},
			},
		},
		{
			Name:        "describe_pod",
			Description: "Get detailed information about a Kubernetes pod",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"namespace": map[string]any{"type": "string"},
					"pod":       map[string]any{"type": "string"},
				},
				"required": []string{"namespace", "pod"},
			},
		},
		{
			Name:        "get_node_status",
			Description: "Get status of Kubernetes nodes",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"node": map[string]any{"type": "string"},
				},
			},
		},
	}, nil
}


