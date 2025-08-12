package calico

import "time"

// InstallOptions contains configuration options for Calico installation.
// These options control how Calico is installed and configured in the cluster.
type InstallOptions struct {
	Version          string         // Calico version to install
	InstallationType string         // Installation method: "operator" or "manifest"
	CNI              string         // CNI type: "calico", "flannel", or "none"
	IPAM             string         // IPAM type: "calico-ipam" or "host-local"
	Datastore        string         // Datastore type: "kubernetes" or "etcd"
	FelixLogSeverity string         // Felix logging severity level
	TyphaLogSeverity string         // Typha logging severity level
	OperatorLogLevel string         // Operator logging level
	EnablePrometheus bool           // Whether to enable Prometheus metrics
	EnableTigera     bool           // Whether to enable Tigera features
	CustomManifests  []string       // Additional custom manifest files
	CustomConfig     map[string]any // Additional custom configuration
}

// UpgradeOptions contains configuration options for Calico upgrades.
// These options control the upgrade process and validation.
type UpgradeOptions struct {
	Version              string // Target Calico version for upgrade
	BackupBeforeUpgrade  bool   // Whether to backup configuration before upgrade
	ValidateAfterUpgrade bool   // Whether to validate the upgrade after completion
}

// UninstallOptions contains configuration options for Calico uninstallation.
// These options control what resources are removed during uninstallation.
type UninstallOptions struct {
	RemoveCRDs       bool // Whether to remove CustomResourceDefinitions
	RemoveNamespace  bool // Whether to remove the Calico namespace
	RemoveFinalizers bool // Whether to remove finalizers from resources
}

// CalicoStatus represents the overall status of Calico installation and components.
// This provides a comprehensive view of the Calico system health and state.
type CalicoStatus struct {
	Installed       bool                       // Whether Calico is installed
	Version         string                     // Current Calico version
	Components      map[string]ComponentStatus // Status of individual components
	Nodes           []NodeStatus               // Status of Calico nodes
	IPPools         []IPPoolStatus             // Status of IP pools
	BGPPeers        []BGPPeerStatus            // Status of BGP peers
	NetworkPolicies []PolicyStatus             // Status of network policies
	LastUpdated     time.Time                  // When the status was last updated
}

// ComponentStatus represents the status of a specific Calico component.
// This tracks the health and readiness of individual system components.
type ComponentStatus struct {
	Name    string // Component name (e.g., "Felix", "Typha", "BGP")
	Healthy bool   // Whether the component is healthy
	Ready   bool   // Whether the component is ready to serve traffic
	Message string // Status message or error description
	Version string // Component version
}

// NodeStatus represents the status of a Calico node in the cluster.
// This tracks the health and configuration of individual nodes.
type NodeStatus struct {
	Name  string // Node name
	Ready bool   // Whether the node is ready
	BGP   bool   // Whether BGP is enabled on the node
	Felix bool   // Whether Felix is running on the node
	Typha bool   // Whether Typha is running on the node
	IP    string // Node IP address
	ASN   int    // BGP ASN for the node
}

// IPPoolStatus represents the status of an IP pool configuration.
// This tracks the configuration and health of IP address pools.
type IPPoolStatus struct {
	Name      string // IP pool name
	CIDR      string // IP pool CIDR block
	BlockSize int    // Block size for IP allocation
	IPIP      bool   // Whether IPIP encapsulation is enabled
	VXLAN     bool   // Whether VXLAN encapsulation is enabled
	Disabled  bool   // Whether the pool is disabled
}

// BGPPeerStatus represents the status of a BGP peer connection.
// This tracks the health and state of BGP peering relationships.
type BGPPeerStatus struct {
	Name   string        // BGP peer name
	ASN    int           // BGP peer ASN
	IP     string        // BGP peer IP address
	State  string        // BGP session state
	Uptime time.Duration // How long the session has been established
}

// PolicyStatus represents the status of a network policy.
// This tracks whether policies are properly applied and enforced.
type PolicyStatus struct {
	Name      string // Policy name
	Namespace string // Policy namespace
	Type      string // Policy type (e.g., "NetworkPolicy", "GlobalNetworkPolicy")
	Applied   bool   // Whether the policy is applied to the cluster
}

// IPPool represents a Calico IP pool configuration.
// IP pools define ranges of IP addresses that can be assigned to pods.
type IPPool struct {
	Name        string            // IP pool name
	CIDR        string            // IP range in CIDR notation
	BlockSize   int               // Size of IP blocks for allocation
	IPIPMode    string            // IPIP encapsulation mode: "Always", "CrossSubnet", "Never"
	VXLANMode   string            // VXLAN encapsulation mode: "Always", "CrossSubnet", "Never"
	Disabled    bool              // Whether the pool is disabled
	Annotations map[string]string // Additional metadata annotations
}

// NetworkPolicy represents a Calico network policy.
// Network policies control traffic flow between pods and external endpoints.
type NetworkPolicy struct {
	Name         string            // Policy name
	Namespace    string            // Policy namespace
	Selector     string            // Pod selector for policy application
	IngressRules []Rule            // Incoming traffic rules
	EgressRules  []Rule            // Outgoing traffic rules
	Types        []string          // Policy types: "Ingress", "Egress"
	Annotations  map[string]string // Additional metadata annotations
}

// Rule represents a network policy rule for controlling traffic.
// Rules define what traffic is allowed, denied, or logged.
type Rule struct {
	Action      string       // Rule action: "Allow", "Deny", "Log"
	Protocol    string       // Network protocol (e.g., "TCP", "UDP", "ICMP")
	Source      RuleEndpoint // Source endpoint specification
	Destination RuleEndpoint // Destination endpoint specification
	HTTP        *HTTPMatch   // HTTP-specific matching criteria
	ICMP        *ICMPMatch   // ICMP-specific matching criteria
}

// RuleEndpoint represents the source or destination of a network policy rule.
// This defines which endpoints the rule applies to.
type RuleEndpoint struct {
	Nets        []string // Network CIDR blocks
	NotNets     []string // Excluded network CIDR blocks
	Selector    string   // Kubernetes label selector
	NotSelector string   // Excluded Kubernetes label selector
	Ports       []Port   // Port specifications
	NotPorts    []Port   // Excluded port specifications
}

// Port represents a port specification for network policy rules.
// This defines which ports the rule applies to.
type Port struct {
	Number   int    // Port number
	Protocol string // Protocol (e.g., "TCP", "UDP")
	EndPort  int    // End port for port ranges
}

// HTTPMatch represents HTTP-specific matching criteria for network policies.
// This allows fine-grained control over HTTP traffic.
type HTTPMatch struct {
	Methods []string          // HTTP methods to match
	Paths   []string          // HTTP paths to match
	Headers map[string]string // HTTP headers to match
}

// ICMPMatch represents ICMP-specific matching criteria for network policies.
// This allows control over ICMP traffic types and codes.
type ICMPMatch struct {
	Type int // ICMP type
	Code int // ICMP code
}

// BGPPeer represents a BGP peer configuration.
// BGP peers enable routing information exchange with external networks.
type BGPPeer struct {
	Name         string            // BGP peer name
	ASN          int               // BGP peer ASN
	IP           string            // BGP peer IP address
	NodeSelector string            // Node selector for peer placement
	Password     string            // BGP authentication password
	Annotations  map[string]string // Additional metadata annotations
}

// HealthCheckResult represents the result of a comprehensive health check.
// This provides an overview of system health across all components.
type HealthCheckResult struct {
	Overall     bool            // Overall system health status
	Components  map[string]bool // Health status of individual components
	Nodes       map[string]bool // Health status of individual nodes
	BGP         map[string]bool // Health status of BGP sessions
	LastChecked time.Time       // When the health check was performed
	Errors      []string        // List of errors found during health check
}

// FelixConfiguration represents Felix configuration settings.
// Felix is the Calico agent that runs on each node and handles networking.
type FelixConfiguration struct {
	Name                            string // Configuration name
	LogSeverityScreen               string // Log severity for console output
	LogSeverityFile                 string // Log severity for file output
	LogSeveritySys                  string // Log severity for syslog
	PrometheusMetricsEnabled        bool   // Whether to enable Prometheus metrics
	PrometheusMetricsPort           int    // Port for Prometheus metrics endpoint
	PrometheusGoMetricsEnabled      bool   // Whether to enable Go runtime metrics
	PrometheusProcessMetricsEnabled bool   // Whether to enable process metrics
}

// BGPConfiguration represents BGP configuration settings.
// This controls BGP routing behavior across the cluster.
type BGPConfiguration struct {
	Name                   string   // Configuration name
	ASNumber               int      // Autonomous System Number for the cluster
	ServiceClusterIPs      []string // Service cluster IP ranges to advertise
	ServiceExternalIPs     []string // Service external IP ranges to advertise
	ServiceLoadBalancerIPs []string // Service load balancer IP ranges to advertise
}

// ClusterInformation represents general cluster information.
// This provides metadata about the cluster and Calico installation.
type ClusterInformation struct {
	Name          string // Cluster name
	CalicoVersion string // Installed Calico version
	ClusterType   string // Type of cluster (e.g., "k8s", "openshift")
	DatastoreType string // Datastore type (e.g., "kubernetes", "etcd")
}

// GlobalNetworkSet represents a global network set configuration.
// Global network sets define IP ranges that can be referenced in policies.
type GlobalNetworkSet struct {
	Name        string            // Network set name
	Nets        []string          // IP ranges in CIDR notation
	Labels      map[string]string // Labels for identification
	Annotations map[string]string // Additional metadata annotations
}

// HostEndpoint represents a host endpoint configuration.
// Host endpoints allow policies to be applied to host interfaces.
type HostEndpoint struct {
	Name        string            // Endpoint name
	Node        string            // Node name where the endpoint exists
	Interface   string            // Network interface name
	ExpectedIPs []string          // Expected IP addresses on the interface
	Profiles    []string          // Security profiles to apply
	Labels      map[string]string // Labels for identification
	Annotations map[string]string // Additional metadata annotations
}

// WorkloadEndpoint represents a workload endpoint configuration.
// Workload endpoints represent pod network interfaces.
type WorkloadEndpoint struct {
	Name        string            // Endpoint name
	Namespace   string            // Pod namespace
	Pod         string            // Pod name
	Container   string            // Container ID
	Interface   string            // Network interface name
	ExpectedIPs []string          // Expected IP addresses on the interface
	Profiles    []string          // Security profiles to apply
	Labels      map[string]string // Labels for identification
	Annotations map[string]string // Additional metadata annotations
}
