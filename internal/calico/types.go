package calico

import "time"

// InstallOptions contains options for Calico installation
type InstallOptions struct {
	Version          string
	InstallationType string // "operator", "manifest"
	CNI              string // "calico", "flannel", "none"
	IPAM             string // "calico-ipam", "host-local"
	Datastore        string // "kubernetes", "etcd"
	FelixLogSeverity string
	TyphaLogSeverity string
	OperatorLogLevel string
	EnablePrometheus bool
	EnableTigera     bool
	CustomManifests  []string
	CustomConfig     map[string]any
}

// UpgradeOptions contains options for Calico upgrade
type UpgradeOptions struct {
	Version              string
	BackupBeforeUpgrade  bool
	ValidateAfterUpgrade bool
}

// UninstallOptions contains options for Calico uninstallation
type UninstallOptions struct {
	RemoveCRDs       bool
	RemoveNamespace  bool
	RemoveFinalizers bool
}

// CalicoStatus represents the status of Calico installation
type CalicoStatus struct {
	Installed       bool
	Version         string
	Components      map[string]ComponentStatus
	Nodes           []NodeStatus
	IPPools         []IPPoolStatus
	BGPPeers        []BGPPeerStatus
	NetworkPolicies []PolicyStatus
	LastUpdated     time.Time
}

// ComponentStatus represents the status of a Calico component
type ComponentStatus struct {
	Name    string
	Healthy bool
	Ready   bool
	Message string
	Version string
}

// NodeStatus represents the status of a Calico node
type NodeStatus struct {
	Name  string
	Ready bool
	BGP   bool
	Felix bool
	Typha bool
	IP    string
	ASN   int
}

// IPPoolStatus represents the status of an IP pool
type IPPoolStatus struct {
	Name      string
	CIDR      string
	BlockSize int
	IPIP      bool
	VXLAN     bool
	Disabled  bool
}

// BGPPeerStatus represents the status of a BGP peer
type BGPPeerStatus struct {
	Name   string
	ASN    int
	IP     string
	State  string
	Uptime time.Duration
}

// PolicyStatus represents the status of a network policy
type PolicyStatus struct {
	Name      string
	Namespace string
	Type      string
	Applied   bool
}

// IPPool represents a Calico IP pool
type IPPool struct {
	Name        string
	CIDR        string
	BlockSize   int
	IPIPMode    string // "Always", "CrossSubnet", "Never"
	VXLANMode   string // "Always", "CrossSubnet", "Never"
	Disabled    bool
	Annotations map[string]string
}

// NetworkPolicy represents a Calico network policy
type NetworkPolicy struct {
	Name         string
	Namespace    string
	Selector     string
	IngressRules []Rule
	EgressRules  []Rule
	Types        []string // "Ingress", "Egress"
	Annotations  map[string]string
}

// Rule represents a network policy rule
type Rule struct {
	Action      string // "Allow", "Deny", "Log"
	Protocol    string
	Source      RuleEndpoint
	Destination RuleEndpoint
	HTTP        *HTTPMatch
	ICMP        *ICMPMatch
}

// RuleEndpoint represents a rule endpoint
type RuleEndpoint struct {
	Nets        []string
	NotNets     []string
	Selector    string
	NotSelector string
	Ports       []Port
	NotPorts    []Port
}

// Port represents a port specification
type Port struct {
	Number   int
	Protocol string
	EndPort  int
}

// HTTPMatch represents HTTP match criteria
type HTTPMatch struct {
	Methods []string
	Paths   []string
	Headers map[string]string
}

// ICMPMatch represents ICMP match criteria
type ICMPMatch struct {
	Type int
	Code int
}

// BGPPeer represents a BGP peer
type BGPPeer struct {
	Name         string
	ASN          int
	IP           string
	NodeSelector string
	Password     string
	Annotations  map[string]string
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Overall     bool
	Components  map[string]bool
	Nodes       map[string]bool
	BGP         map[string]bool
	LastChecked time.Time
	Errors      []string
}

// FelixConfiguration represents Felix configuration
type FelixConfiguration struct {
	Name                            string
	LogSeverityScreen               string
	LogSeverityFile                 string
	LogSeveritySys                  string
	PrometheusMetricsEnabled        bool
	PrometheusMetricsPort           int
	PrometheusGoMetricsEnabled      bool
	PrometheusProcessMetricsEnabled bool
}

// BGPConfiguration represents BGP configuration
type BGPConfiguration struct {
	Name                   string
	ASNumber               int
	ServiceClusterIPs      []string
	ServiceExternalIPs     []string
	ServiceLoadBalancerIPs []string
}

// ClusterInformation represents cluster information
type ClusterInformation struct {
	Name          string
	CalicoVersion string
	ClusterType   string
	DatastoreType string
}

// GlobalNetworkSet represents a global network set
type GlobalNetworkSet struct {
	Name        string
	Nets        []string
	Labels      map[string]string
	Annotations map[string]string
}

// HostEndpoint represents a host endpoint
type HostEndpoint struct {
	Name        string
	Node        string
	Interface   string
	ExpectedIPs []string
	Profiles    []string
	Labels      map[string]string
	Annotations map[string]string
}

// WorkloadEndpoint represents a workload endpoint
type WorkloadEndpoint struct {
	Name        string
	Namespace   string
	Pod         string
	Container   string
	Interface   string
	ExpectedIPs []string
	Profiles    []string
	Labels      map[string]string
	Annotations map[string]string
}
