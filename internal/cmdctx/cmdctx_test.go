package cmdctx

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestNewCmdCtx(t *testing.T) {
	tests := []struct {
		name        string
		kubeConfig  string
		kubeContext string
		wantConfig  string
		wantContext string
	}{
		{
			name:        "with kubeconfig and context",
			kubeConfig:  "/path/to/kubeconfig",
			kubeContext: "test-context",
			wantConfig:  "/path/to/kubeconfig",
			wantContext: "test-context",
		},
		{
			name:        "with kubeconfig only",
			kubeConfig:  "/path/to/kubeconfig",
			kubeContext: "",
			wantConfig:  "/path/to/kubeconfig",
			wantContext: "",
		},
		{
			name:        "empty values",
			kubeConfig:  "",
			kubeContext: "",
			wantConfig:  "",
			wantContext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewCmdCtx(tt.kubeConfig, tt.kubeContext)
			if ctx.kubeConfig != tt.wantConfig {
				t.Errorf("NewCmdCtx().kubeConfig = %v, want %v", ctx.kubeConfig, tt.wantConfig)
			}
			if ctx.kubeContext != tt.wantContext {
				t.Errorf("NewCmdCtx().kubeContext = %v, want %v", ctx.kubeContext, tt.wantContext)
			}
		})
	}
}

func TestToContext(t *testing.T) {
	ctx := NewCmdCtx("", "")
	parentCtx := context.Background()

	// Test context creation
	newCtx := ctx.ToContext(parentCtx)
	if newCtx == nil {
		t.Error("ToContext() returned nil context")
	}

	// Test context value retrieval
	retrievedCtx := CmdCtxFromContext(newCtx)
	if retrievedCtx != ctx {
		t.Error("CmdCtxFromContext() returned different context")
	}

	// Test cancellation
	ctx.Cancel()
	select {
	case <-newCtx.Done():
		// Context was cancelled as expected
	default:
		t.Error("Context was not cancelled")
	}
}

func TestGetK8sConfig(t *testing.T) {
	// Create a temporary kubeconfig file
	tmpDir, err := os.MkdirTemp("", "kubeconfig-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
	config := api.Config{
		Clusters: map[string]*api.Cluster{
			"test-cluster": {
				Server: "https://test-server:6443",
			},
		},
		Contexts: map[string]*api.Context{
			"test-context": {
				Cluster:  "test-cluster",
				AuthInfo: "test-user",
			},
		},
		CurrentContext: "test-context",
	}

	err = clientcmd.WriteToFile(config, kubeconfigPath)
	if err != nil {
		t.Fatalf("Failed to write kubeconfig: %v", err)
	}

	tests := []struct {
		name        string
		kubeConfig  string
		kubeContext string
		wantErr     bool
	}{
		{
			name:        "valid kubeconfig and context",
			kubeConfig:  kubeconfigPath,
			kubeContext: "test-context",
			wantErr:     false,
		},
		{
			name:        "invalid kubeconfig path",
			kubeConfig:  "/non/existent/path",
			kubeContext: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewCmdCtx(tt.kubeConfig, tt.kubeContext)

			if tt.wantErr {
				defer func() {
					if r := recover(); r == nil {
						t.Error("GetK8sConfig() did not panic with invalid config")
					}
				}()
			}

			config := ctx.GetK8sConfig()
			if !tt.wantErr && config == nil {
				t.Error("GetK8sConfig() returned nil config")
			}
		})
	}
}

func TestClientDyn(t *testing.T) {
	ctx := &CmdCtx{
		k8scfg: &rest.Config{},
		dc:     nil,
	}

	// Test client creation
	client := ctx.ClientDyn()
	if client == nil {
		t.Error("ClientDyn() returned nil client")
	}

	// Test client caching
	client2 := ctx.ClientDyn()
	if client != client2 {
		t.Error("ClientDyn() returned different clients")
	}
}

func TestClientset(t *testing.T) {
	ctx := &CmdCtx{
		k8scfg: &rest.Config{},
		cs:     nil,
	}

	// Test clientset creation
	clientset := ctx.Clientset()
	if clientset == nil {
		t.Error("Clientset() returned nil clientset")
	}

	// Test clientset caching
	clientset2 := ctx.Clientset()
	if clientset != clientset2 {
		t.Error("Clientset() returned different clientsets")
	}
}

// func TestContextHelpers(t *testing.T) {
// 	ctx := NewCmdCtx("", "")
// 	parentCtx := context.Background()
// 	newCtx := ctx.ToContext(parentCtx)
//
// 	// Test K8sClientDynFromContext
// 	dynClient := K8sClientDynFromContext(newCtx)
// 	if dynClient == nil {
// 		t.Error("K8sClientDynFromContext() returned nil client")
// 	}
//
// 	// Test K8sClientsetFromContext
// 	clientset := K8sClientsetFromContext(newCtx)
// 	if clientset == nil {
// 		t.Error("K8sClientsetFromContext() returned nil clientset")
// 	}
//
// 	// Test K8sConfigFromContext
// 	config := K8sConfigFromContext(newCtx)
// 	if config == nil {
// 		t.Error("K8sConfigFromContext() returned nil config")
// 	}
//
// 	// Test with invalid context
// 	invalidCtx := context.Background()
// 	func() {
// 		defer func() {
// 			if r := recover(); r == nil {
// 				t.Error("Context helpers did not panic with invalid context")
// 			}
// 		}()
// 		K8sClientDynFromContext(invalidCtx)
// 	}()
// }

func TestCancel(t *testing.T) {
	ctx := NewCmdCtx("", "")
	parentCtx := context.Background()
	newCtx := ctx.ToContext(parentCtx)

	// Test cancellation
	ctx.Cancel()
	select {
	case <-newCtx.Done():
		// Context was cancelled as expected
	default:
		t.Error("Context was not cancelled")
	}

	// Test cancellation without context
	ctx2 := NewCmdCtx("", "")
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Cancel() did not panic without context")
			}
		}()
		ctx2.Cancel()
	}()
}
