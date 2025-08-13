# MCP Server Package

This package implements a Model Context Protocol (MCP) server for Kubernetes and Calico resources and tools.

## Overview

The MCP server provides a standardized interface for AI models and other clients to interact with Kubernetes and Calico resources. It exposes:

- **Resources**: Kubernetes and Calico objects that can be queried and retrieved
- **Tools**: Functions that can be called to perform operations on the cluster

## Architecture

```
┌─────────────────┐    HTTP/REST    ┌─────────────────┐
│   MCP Client    │ ◄──────────────► │   MCP Server    │
└─────────────────┘                 └─────────────────┘
                                              │
                                              ▼
                                    ┌─────────────────┐
                                    │   Providers     │
                                    │                 │
                                    │ ┌─────────────┐ │
                                    │ │ Kubernetes │ │
                                    │ │  Provider  │ │
                                    │ └─────────────┘ │
                                    │                 │
                                    │ ┌─────────────┐ │
                                    │ │   Calico   │ │
                                    │ │  Provider  │ │
                                    │ └─────────────┘ │
                                    └─────────────────┘
                                              │
                                              ▼
                                    ┌─────────────────┐
                                    │   Tools        │
                                    │                 │
                                    │ ┌─────────────┐ │
                                    │ │ Kubernetes │ │
                                    │ │   Tools    │ │
                                    │ └─────────────┘ │
                                    │                 │
                                    │ ┌─────────────┐ │
                                    │ │   Calico   │ │
                                    │ │   Tools    │ │
                                    │ └─────────────┘ │
                                    └─────────────────┘
```

## Features

### Kubernetes Resources
- **Pods**: Container instances running in the cluster
- **Services**: Network endpoints for pods
- **Deployments**: Replica sets of pods
- **Nodes**: Worker machines in the cluster

### Calico Resources
- **Network Policies**: Security rules for network traffic
- **IP Pools**: IP address ranges for pods
- **Endpoints**: Network interfaces for pods

### Kubernetes Tools
- `get_pod_logs`: Retrieve logs from pods
- `describe_pod`: Get detailed pod information
- `get_node_status`: Check node health and status

### Calico Tools
- `get_calico_policies`: List network policies
- `get_calico_ip_pools`: List IP pools
- `get_calico_endpoints`: List network endpoints

## API Endpoints

### Resources
- `GET /mcp/resources` - List all available resources
- `GET /mcp/resources/{uri}` - Get a specific resource by URI

### Tools
- `GET /mcp/tools` - List all available tools
- `POST /mcp/tools/{name}` - Execute a specific tool

### Health
- `GET /health` - Health check endpoint

## Resource URIs

### Kubernetes Resources
- `k8s://pods/{namespace}/{name}` - Pod resource
- `k8s://services/{namespace}/{name}` - Service resource
- `k8s://deployments/{namespace}/{name}` - Deployment resource
- `k8s://nodes/{name}` - Node resource

### Calico Resources
- `calico://networkpolicies/{namespace}/{name}` - Network policy
- `calico://ippools/{name}` - IP pool
- `calico://endpoints/{name}` - Endpoint

## Usage

### Starting the Server

```bash
# Basic usage
clyde mcp

# Custom port
clyde mcp --port 9090

# With kubeconfig
clyde mcp --kubeconfig /path/to/kubeconfig

# With TLS
clyde mcp --enable-tls --tls-cert cert.pem --tls-key key.pem
```

### Environment Variables

- `MCP_PORT`: Server port (default: 8080)
- `MCP_HOST`: Server host (default: 0.0.0.0)
- `MCP_KUBECONFIG_PATH`: Path to kubeconfig file
- `MCP_CALICO_NAMESPACE`: Calico namespace (default: kube-system)
- `MCP_LOG_LEVEL`: Log level (default: info)
- `MCP_ENABLE_TLS`: Enable TLS (default: false)
- `MCP_TLS_CERT_PATH`: TLS certificate path
- `MCP_TLS_KEY_PATH`: TLS private key path

### Example API Calls

#### List Resources
```bash
curl http://localhost:8080/mcp/resources
```

#### Get Kubernetes Pod
```bash
curl http://localhost:8080/mcp/resources/k8s://pods/default/my-pod
```

#### List Tools
```bash
curl http://localhost:8080/mcp/tools
```

#### Execute Tool
```bash
curl -X POST http://localhost:8080/mcp/tools/get_pod_logs \
  -H "Content-Type: application/json" \
  -d '{"namespace": "default", "pod": "my-pod"}'
```

## Configuration

The server can be configured through:

1. **Command-line flags**: Direct configuration when starting
2. **Environment variables**: System-wide or user-specific settings
3. **Configuration file**: Future enhancement for file-based config

## Security

- **TLS Support**: Optional TLS encryption for secure communication
- **Authentication**: Future enhancement for client authentication
- **Authorization**: Future enhancement for role-based access control

## Development

### Adding New Resources

1. Implement the `ResourceProvider` interface
2. Add resource listing and retrieval logic
3. Update the server to use the new provider

### Adding New Tools

1. Implement the `ToolProvider` interface
2. Add tool execution logic
3. Update the server to use the new provider

### Testing

```bash
# Run tests
go test ./internal/mcp/...

# Run with coverage
go test -cover ./internal/mcp/...
```

## Dependencies

- `k8s.io/client-go`: Kubernetes client library
- `k8s.io/apimachinery`: Kubernetes API types
- Standard library: `net/http`, `context`, `encoding/json`

## Future Enhancements

- [ ] gRPC support for better performance
- [ ] WebSocket support for real-time updates
- [ ] Resource streaming for large datasets
- [ ] Authentication and authorization
- [ ] Metrics and monitoring
- [ ] Configuration file support
- [ ] Plugin system for custom providers
- [ ] Resource caching and optimization
- [ ] Multi-cluster support
- [ ] Resource validation and schema enforcement
