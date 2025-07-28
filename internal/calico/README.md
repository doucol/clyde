# Calico Package

The Calico package provides comprehensive Calico OSS management capabilities including installation, configuration, and management of Calico resources as described in the [Calico open source documentation](https://docs.tigera.io/calico).

## Features

### Installation and Management
- **Installation**: Install Calico using operator-based or manifest-based methods
- **Upgrade**: Upgrade Calico to newer versions with backup and validation options
- **Uninstallation**: Complete removal of Calico with configurable cleanup options
- **Status Monitoring**: Get comprehensive status of Calico installation and components

### Configuration Management
- **Felix Configuration**: Get and update Felix component settings
- **BGP Configuration**: Manage BGP routing configuration
- **Cluster Information**: Retrieve cluster and Calico version information

### Resource Management
- **IP Pools**: Create, list, and delete IP pools for pod networking
- **Network Policies**: Manage Calico network policies with complex rule definitions
- **BGP Peers**: Configure BGP peering for external routing
- **Global Network Sets**: Manage global network sets for policy enforcement
- **Host Endpoints**: Configure host-level network endpoints
- **Workload Endpoints**: Manage pod-level network endpoints

### Health and Monitoring
- **Health Checks**: Comprehensive health checking of all Calico components
- **Status Reporting**: Detailed status of nodes, components, and resources
- **Ready State Monitoring**: Wait for Calico to be fully operational

### Advanced Features
- **Custom YAML Application**: Apply custom Calico resources from YAML
- **Manifest Generation**: Generate YAML manifests for all resource types
- **Rule Parsing**: Parse and generate complex network policy rules
- **Logging**: Integrated logging with configurable output

## Usage

### Basic Setup

```go
import "github.com/doucol/clyde/internal/calico"

// Create Kubernetes clients
config, err := rest.InClusterConfig()
if err != nil {
    log.Fatalf("Failed to get cluster config: %v", err)
}

clientset, err := kubernetes.NewForConfig(config)
if err != nil {
    log.Fatalf("Failed to create clientset: %v", err)
}

dynamicClient, err := dynamic.NewForConfig(config)
if err != nil {
    log.Fatalf("Failed to create dynamic client: %v", err)
}

// Create Calico manager
calicoManager := calico.NewCalicoManager(clientset, dynamicClient, config, os.Stdout)
```

### Installation

```go
installOptions := &calico.InstallOptions{
    Version:          "v3.26.1",
    InstallationType: "operator",
    CNI:              "calico",
    IPAM:             "calico-ipam",
    Datastore:        "kubernetes",
    EnablePrometheus: true,
}

err := calicoManager.Install(ctx, installOptions)
if err != nil {
    log.Printf("Failed to install Calico: %v", err)
}
```

### IP Pool Management

```go
// Create IP pool
ipPool := &calico.IPPool{
    Name:      "example-pool",
    CIDR:      "192.168.0.0/16",
    BlockSize: 26,
    IPIPMode:  "CrossSubnet",
    VXLANMode: "Never",
}

err := calicoManager.CreateIPPool(ctx, ipPool)
if err != nil {
    log.Printf("Failed to create IP pool: %v", err)
}

// List IP pools
pools, err := calicoManager.GetIPPools(ctx)
if err != nil {
    log.Printf("Failed to get IP pools: %v", err)
}

// Delete IP pool
err = calicoManager.DeleteIPPool(ctx, "example-pool")
if err != nil {
    log.Printf("Failed to delete IP pool: %v", err)
}
```

### Network Policy Management

```go
// Create network policy
networkPolicy := &calico.NetworkPolicy{
    Name:      "example-policy",
    Namespace: "default",
    Selector:  "app == 'web'",
    Types:     []string{"Ingress", "Egress"},
    IngressRules: []calico.Rule{
        {
            Action:   "Allow",
            Protocol: "TCP",
            Source: calico.RuleEndpoint{
                Selector: "app == 'frontend'",
            },
            Destination: calico.RuleEndpoint{
                Ports: []calico.Port{
                    {Number: 80, Protocol: "TCP"},
                },
            },
        },
    },
}

err := calicoManager.CreateNetworkPolicy(ctx, networkPolicy)
if err != nil {
    log.Printf("Failed to create network policy: %v", err)
}
```

### Configuration Management

```go
// Get Felix configuration
felixConfig, err := calicoManager.GetFelixConfiguration(ctx)
if err != nil {
    log.Printf("Failed to get Felix configuration: %v", err)
}

// Update Felix configuration
if felixConfig != nil {
    felixConfig.LogSeverityScreen = "Info"
    felixConfig.PrometheusMetricsEnabled = true
    felixConfig.PrometheusMetricsPort = 9091

    err = calicoManager.UpdateFelixConfiguration(ctx, felixConfig)
    if err != nil {
        log.Printf("Failed to update Felix configuration: %v", err)
    }
}

// Get BGP configuration
bgpConfig, err := calicoManager.GetBGPConfiguration(ctx)
if err != nil {
    log.Printf("Failed to get BGP configuration: %v", err)
}
```

### Health Monitoring

```go
// Wait for Calico to be ready
err := calicoManager.WaitForReady(ctx, 5*time.Minute)
if err != nil {
    log.Printf("Failed to wait for Calico ready: %v", err)
}

// Get comprehensive status
status, err := calicoManager.GetStatus(ctx)
if err != nil {
    log.Printf("Failed to get status: %v", err)
} else {
    fmt.Printf("Calico installed: %v, Version: %s\n", status.Installed, status.Version)
}

// Perform health check
healthResult, err := calicoManager.HealthCheck(ctx)
if err != nil {
    log.Printf("Failed to perform health check: %v", err)
} else {
    fmt.Printf("Calico health: %v, Errors: %d\n", healthResult.Overall, len(healthResult.Errors))
}
```

### Upgrade and Uninstall

```go
// Upgrade Calico
upgradeOptions := &calico.UpgradeOptions{
    Version:              "v3.27.0",
    BackupBeforeUpgrade:  true,
    ValidateAfterUpgrade: true,
}

err := calicoManager.Upgrade(ctx, upgradeOptions)
if err != nil {
    log.Printf("Failed to upgrade Calico: %v", err)
}

// Uninstall Calico
uninstallOptions := &calico.UninstallOptions{
    RemoveCRDs:      true,
    RemoveNamespace:  true,
    RemoveFinalizers: true,
}

err = calicoManager.Uninstall(ctx, uninstallOptions)
if err != nil {
    log.Printf("Failed to uninstall Calico: %v", err)
}
```

### Custom Resource Application

```go
// Apply custom YAML resource
customYAML := `apiVersion: crd.projectcalico.org/v1
kind: IPPool
metadata:
  name: custom-pool
spec:
  cidr: 10.0.0.0/16
  blockSize: 26
  ipipMode: CrossSubnet
  vxlanMode: Never
  natOutgoing: true
  nodeSelector: all()
`

err := calicoManager.ApplyResource(ctx, customYAML)
if err != nil {
    log.Printf("Failed to apply custom resource: %v", err)
}
```

## Resource Types

### IPPool
Represents a Calico IP pool for pod networking:
- `Name`: Pool identifier
- `CIDR`: IP range (e.g., "192.168.0.0/16")
- `BlockSize`: Block size for IP allocation
- `IPIPMode`: IPIP encapsulation mode
- `VXLANMode`: VXLAN encapsulation mode
- `Disabled`: Whether the pool is disabled

### NetworkPolicy
Represents a Calico network policy:
- `Name`: Policy name
- `Namespace`: Target namespace
- `Selector`: Pod selector
- `Types`: Policy types (Ingress/Egress)
- `IngressRules`: Incoming traffic rules
- `EgressRules`: Outgoing traffic rules

### Rule
Represents a network policy rule:
- `Action`: Allow/Deny/Log
- `Protocol`: Network protocol
- `Source`: Source endpoint specification
- `Destination`: Destination endpoint specification
- `HTTP`: HTTP-specific matching
- `ICMP`: ICMP-specific matching

### BGPPeer
Represents a BGP peer configuration:
- `Name`: Peer identifier
- `ASN`: Autonomous System Number
- `IP`: Peer IP address
- `NodeSelector`: Node selector for peer
- `Password`: BGP password (optional)

### FelixConfiguration
Represents Felix component configuration:
- `LogSeverityScreen`: Screen log level
- `LogSeverityFile`: File log level
- `LogSeveritySys`: Syslog level
- `PrometheusMetricsEnabled`: Enable Prometheus metrics
- `PrometheusMetricsPort`: Metrics port
- `PrometheusGoMetricsEnabled`: Enable Go metrics
- `PrometheusProcessMetricsEnabled`: Enable process metrics

### BGPConfiguration
Represents BGP routing configuration:
- `ASNumber`: Autonomous System Number
- `ServiceClusterIPs`: Service cluster IP ranges
- `ServiceExternalIPs`: Service external IP ranges
- `ServiceLoadBalancerIPs`: Load balancer IP ranges

## Error Handling

The package provides comprehensive error handling with detailed error messages. All methods return errors that should be checked and handled appropriately:

```go
err := calicoManager.Install(ctx, options)
if err != nil {
    log.Printf("Installation failed: %v", err)
    // Handle error appropriately
}
```

## Logging

The Calico manager includes built-in logging that can be configured through the logger parameter:

```go
// Use stdout for logging
calicoManager := calico.NewCalicoManager(clientset, dynamicClient, config, os.Stdout)

// Use custom logger
var buf bytes.Buffer
calicoManager := calico.NewCalicoManager(clientset, dynamicClient, config, &buf)
```

## Dependencies

The package requires the following Kubernetes client-go dependencies:
- `k8s.io/client-go/kubernetes`
- `k8s.io/client-go/dynamic`
- `k8s.io/client-go/rest`
- `k8s.io/apimachinery/pkg/apis/meta/v1`
- `k8s.io/apimachinery/pkg/apis/meta/v1/unstructured`
- `k8s.io/apimachinery/pkg/runtime/schema`
- `k8s.io/apimachinery/pkg/util/yaml`

## Examples

See `example_test.go` for comprehensive usage examples covering all major functionality.

## Documentation

For more information about Calico OSS, refer to the official documentation at [docs.tigera.io/calico](https://docs.tigera.io/calico). 