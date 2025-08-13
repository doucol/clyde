package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server represents the MCP server for Kubernetes and Calico
type Server struct {
	httpServer      *http.Server
	port            int
	logger          *slog.Logger
	k8sProvider     *KubernetesProvider
	calicoProvider  *CalicoProvider
	k8sTools        *KubernetesTools
	calicoTools     *CalicoTools
}

// NewServer creates a new MCP server instance
func NewServer(port int, logger *slog.Logger) *Server {
	return &Server{
		port:   port,
		logger: logger,
	}
}

// Start initializes and starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	// Initialize providers and tools
	s.k8sProvider = NewKubernetesProvider(nil) // TODO: Pass actual k8s client
	s.calicoProvider = NewCalicoProvider(nil)  // TODO: Pass actual calico client
	s.k8sTools = NewKubernetesTools(nil)      // TODO: Pass actual k8s client
	s.calicoTools = NewCalicoTools(nil)       // TODO: Pass actual calico client

	// Create HTTP server with MCP endpoints
	mux := http.NewServeMux()
	
	// MCP protocol endpoints
	mux.HandleFunc("/mcp/resources", s.handleListResources)
	mux.HandleFunc("/mcp/resources/", s.handleGetResource)
	mux.HandleFunc("/mcp/tools", s.handleListTools)
	mux.HandleFunc("/mcp/tools/", s.handleCallTool)
	
	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	s.logger.Info("Starting MCP server", "port", s.port)

	// Start server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Failed to serve", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down MCP server...")
	
	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Error during server shutdown", "error", err)
	}

	return nil
}

// Stop gracefully stops the server
func (s *Server) Stop() {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}
}

// handleListResources handles the MCP resources listing endpoint
func (s *Server) handleListResources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Combine resources from both providers
	k8sResources, err := s.k8sProvider.ListResources(r.Context())
	if err != nil {
		s.logger.Error("Failed to list Kubernetes resources", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	calicoResources, err := s.calicoProvider.ListResources(r.Context())
	if err != nil {
		s.logger.Error("Failed to list Calico resources", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	allResources := append(k8sResources, calicoResources...)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"resources": allResources,
		"count":     len(allResources),
	})
}

// handleGetResource handles the MCP resource retrieval endpoint
func (s *Server) handleGetResource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract URI from path
	uri := r.URL.Path[len("/mcp/resources/"):]
	if uri == "" {
		http.Error(w, "Resource URI required", http.StatusBadRequest)
		return
	}

	// Try to get resource from both providers
	var resource *Resource
	var err error

	// Check if it's a Kubernetes resource
	if resource, err = s.k8sProvider.GetResource(r.Context(), uri); err == nil && resource != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resource)
		return
	}

	// Check if it's a Calico resource
	if resource, err = s.calicoProvider.GetResource(r.Context(), uri); err == nil && resource != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resource)
		return
	}

	http.Error(w, "Resource not found", http.StatusNotFound)
}

// handleListTools handles the MCP tools listing endpoint
func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Combine tools from both providers
	k8sTools, err := s.k8sTools.ListTools(r.Context())
	if err != nil {
		s.logger.Error("Failed to list Kubernetes tools", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	calicoTools, err := s.calicoTools.ListTools(r.Context())
	if err != nil {
		s.logger.Error("Failed to list Calico tools", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	allTools := append(k8sTools, calicoTools...)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": allTools,
		"count": len(allTools),
	})
}

// handleCallTool handles the MCP tool execution endpoint
func (s *Server) handleCallTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tool name from path
	toolName := r.URL.Path[len("/mcp/tools/"):]
	if toolName == "" {
		http.Error(w, "Tool name required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var requestBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Try to call tool from both providers
	var result interface{}
	var err error

	// Check if it's a Kubernetes tool
	if result, err = s.k8sTools.CallTool(r.Context(), toolName, requestBody); err == nil && result != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": result,
		})
		return
	}

	// Check if it's a Calico tool
	if result, err = s.calicoTools.CallTool(r.Context(), toolName, requestBody); err == nil && result != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"result": result,
		})
		return
	}

	http.Error(w, "Tool not found or execution failed", http.StatusNotFound)
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}
