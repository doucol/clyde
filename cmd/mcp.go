package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/doucol/clyde/internal/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpPort            int
	mcpHost            string
	mcpKubeconfigPath  string
	mcpCalicoNamespace string
	mcpLogLevel        string
	mcpEnableTLS       bool
	mcpTLSCertPath     string
	mcpTLSKeyPath      string
)

// mcpCmd represents the mcp command
var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MCP server for Kubernetes and Calico",
	Long: `Start the Model Context Protocol (MCP) server that provides
Kubernetes and Calico resources and tools to MCP clients.

The MCP server exposes:
- Kubernetes resources (pods, services, deployments, nodes)
- Calico resources (network policies, IP pools, endpoints)
- Tools for interacting with both systems

Examples:
  clyde mcp --port 8080
  clyde mcp --kubeconfig /path/to/kubeconfig
  clyde mcp --enable-tls --tls-cert cert.pem --tls-key key.pem`,
	RunE: runMCPServer,
}

func init() {
	// MCP server configuration flags
	mcpCmd.Flags().IntVarP(&mcpPort, "port", "p", 8080, "Port for the MCP server to listen on")
	mcpCmd.Flags().StringVarP(&mcpHost, "host", "H", "0.0.0.0", "Host for the MCP server to bind to")
	mcpCmd.Flags().StringVarP(&mcpKubeconfigPath, "kubeconfig", "k", "", "Path to kubeconfig file (default: use in-cluster config)")
	mcpCmd.Flags().StringVarP(&mcpCalicoNamespace, "calico-namespace", "c", "kube-system", "Namespace where Calico resources are deployed")
	mcpCmd.Flags().StringVarP(&mcpLogLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	mcpCmd.Flags().BoolVarP(&mcpEnableTLS, "enable-tls", "t", false, "Enable TLS for the MCP server")
	mcpCmd.Flags().StringVarP(&mcpTLSCertPath, "tls-cert", "", "", "Path to TLS certificate file")
	mcpCmd.Flags().StringVarP(&mcpTLSKeyPath, "tls-key", "", "", "Path to TLS private key file")

	// Mark TLS flags as required when TLS is enabled
	mcpCmd.MarkFlagsRequiredTogether("enable-tls", "tls-cert", "tls-key")
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	// Set up logging
	logLevel := slog.LevelInfo
	switch mcpLogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Create MCP configuration
	config := mcp.DefaultConfig()
	config.Port = mcpPort
	config.Host = mcpHost
	config.KubeconfigPath = mcpKubeconfigPath
	config.CalicoNamespace = mcpCalicoNamespace
	config.LogLevel = mcpLogLevel
	config.EnableTLS = mcpEnableTLS
	config.TLSCertPath = mcpTLSCertPath
	config.TLSKeyPath = mcpTLSKeyPath

	// Load configuration from environment variables
	config.LoadFromEnv()

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	logger.Info("Starting MCP server",
		"port", config.Port,
		"host", config.Host,
		"kubeconfig", config.KubeconfigPath,
		"calico_namespace", config.CalicoNamespace,
		"tls_enabled", config.EnableTLS,
	)

	// Create and start MCP server
	server := mcp.NewServer(config.Port, logger)

	// Set up context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the server
	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	return nil
}
