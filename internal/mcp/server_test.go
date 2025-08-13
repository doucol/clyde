package mcp

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	server := NewServer(8080, nil)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
	if server.port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.port)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if config.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Port)
	}
	if config.Host != "0.0.0.0" {
		t.Errorf("Expected default host 0.0.0.0, got %s", config.Host)
	}
}

func TestConfigValidation(t *testing.T) {
	config := DefaultConfig()
	
	// Test valid config
	if err := config.Validate(); err != nil {
		t.Errorf("Valid config should not return error: %v", err)
	}
	
	// Test invalid port
	config.Port = 0
	if err := config.Validate(); err != ErrInvalidPort {
		t.Errorf("Expected ErrInvalidPort for port 0, got %v", err)
	}
	
	// Test invalid port
	config.Port = 70000
	if err := config.Validate(); err != ErrInvalidPort {
		t.Errorf("Expected ErrInvalidPort for port 70000, got %v", err)
	}
	
	// Reset port and test TLS validation
	config.Port = 8080
	config.EnableTLS = true
	if err := config.Validate(); err != ErrMissingTLSCert {
		t.Errorf("Expected ErrMissingTLSCert when TLS enabled without cert, got %v", err)
	}
	
	config.TLSCertPath = "cert.pem"
	if err := config.Validate(); err != ErrMissingTLSKey {
		t.Errorf("Expected ErrMissingTLSKey when TLS enabled without key, got %v", err)
	}
}

func TestKubernetesProvider(t *testing.T) {
	provider := NewKubernetesProvider(nil)
	if provider == nil {
		t.Fatal("NewKubernetesProvider returned nil")
	}
	
	ctx := context.Background()
	resources, err := provider.ListResources(ctx)
	if err != nil {
		t.Errorf("ListResources should not return error when client is nil: %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("Expected 0 resources when client is nil, got %d", len(resources))
	}
}

func TestCalicoProvider(t *testing.T) {
	provider := NewCalicoProvider(nil)
	if provider == nil {
		t.Fatal("NewCalicoProvider returned nil")
	}
	
	ctx := context.Background()
	resources, err := provider.ListResources(ctx)
	if err != nil {
		t.Errorf("ListResources should not return error when client is nil: %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("Expected 0 resources when client is nil, got %d", len(resources))
	}
}

func TestKubernetesTools(t *testing.T) {
	tools := NewKubernetesTools(nil)
	if tools == nil {
		t.Fatal("NewKubernetesTools returned nil")
	}
	
	ctx := context.Background()
	toolList, err := tools.ListTools(ctx)
	if err != nil {
		t.Errorf("ListTools should not return error when client is nil: %v", err)
	}
	if len(toolList) == 0 {
		t.Errorf("Expected tools to be available even when client is nil")
	}
}

func TestCalicoTools(t *testing.T) {
	tools := NewCalicoTools(nil)
	if tools == nil {
		t.Fatal("NewCalicoTools returned nil")
	}
	
	ctx := context.Background()
	toolList, err := tools.ListTools(ctx)
	if err != nil {
		t.Errorf("ListTools should not return error when client is nil: %v", err)
	}
	if len(toolList) == 0 {
		t.Errorf("Expected tools to be available even when client is nil")
	}
}

func TestServerGracefulShutdown(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := NewServer(0, logger) // Use port 0 for testing
	
	// Start server in goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	go func() {
		server.Start(ctx)
	}()
	
	// Wait a bit for server to start
	time.Sleep(10 * time.Millisecond)
	
	// Stop server
	server.Stop()
	
	// Server should stop gracefully
	time.Sleep(50 * time.Millisecond)
}
