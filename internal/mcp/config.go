package mcp

import (
	"os"
	"strconv"
)

// Config holds the configuration for the MCP server
type Config struct {
	Port            int    `json:"port"`
	Host            string `json:"host"`
	KubeconfigPath  string `json:"kubeconfig_path"`
	CalicoNamespace string `json:"calico_namespace"`
	LogLevel        string `json:"log_level"`
	EnableTLS       bool   `json:"enable_tls"`
	TLSCertPath     string `json:"tls_cert_path"`
	TLSKeyPath      string `json:"tls_key_path"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		Host:            "0.0.0.0",
		KubeconfigPath:  "",
		CalicoNamespace: "kube-system",
		LogLevel:        "info",
		EnableTLS:       false,
		TLSCertPath:     "",
		TLSKeyPath:      "",
	}
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	if port := os.Getenv("MCP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Port = p
		}
	}

	if host := os.Getenv("MCP_HOST"); host != "" {
		c.Host = host
	}

	if kubeconfig := os.Getenv("MCP_KUBECONFIG_PATH"); kubeconfig != "" {
		c.KubeconfigPath = kubeconfig
	}

	if namespace := os.Getenv("MCP_CALICO_NAMESPACE"); namespace != "" {
		c.CalicoNamespace = namespace
	}

	if logLevel := os.Getenv("MCP_LOG_LEVEL"); logLevel != "" {
		c.LogLevel = logLevel
	}

	if enableTLS := os.Getenv("MCP_ENABLE_TLS"); enableTLS != "" {
		if b, err := strconv.ParseBool(enableTLS); err == nil {
			c.EnableTLS = b
		}
	}

	if certPath := os.Getenv("MCP_TLS_CERT_PATH"); certPath != "" {
		c.TLSCertPath = certPath
	}

	if keyPath := os.Getenv("MCP_TLS_KEY_PATH"); keyPath != "" {
		c.TLSKeyPath = keyPath
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return ErrInvalidPort
	}

	if c.EnableTLS {
		if c.TLSCertPath == "" {
			return ErrMissingTLSCert
		}
		if c.TLSKeyPath == "" {
			return ErrMissingTLSKey
		}
	}

	return nil
}
