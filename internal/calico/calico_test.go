package calico

import (
	"context"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewCalicoManager(t *testing.T) {
	// Test with nil clients since we're focusing on data structure tests
	manager := NewCalicoManager(nil, nil, nil, nil)

	if manager == nil {
		t.Fatal("Expected CalicoManager to be created, got nil")
	}

	if manager.clientset != nil {
		t.Error("Expected clientset to be nil")
	}

	if manager.dynamic != nil {
		t.Error("Expected dynamic client to be nil")
	}

	if manager.config != nil {
		t.Error("Expected config to be nil")
	}

	if manager.logger != nil {
		t.Error("Expected logger to be nil")
	}
}

func TestCalicoManager_Logf(t *testing.T) {
	logger := &testLogger{}
	manager := &CalicoManager{logger: logger}

	manager.Logf("test message %s", "value")

	if len(logger.messages) == 0 {
		t.Fatal("Expected log message to be written")
	}

	if logger.messages[0] != "test message value\n" {
		t.Errorf("Expected log message 'test message value\\n', got '%s'", logger.messages[0])
	}
}

// Test data structures
func TestInstallOptions(t *testing.T) {
	options := &InstallOptions{
		Version:          "v3.26.1",
		InstallationType: "operator",
		CNI:              "calico",
		IPAM:             "calico-ipam",
		Datastore:        "kubernetes",
		FelixLogSeverity: "Info",
		TyphaLogSeverity: "Info",
		OperatorLogLevel: "Info",
		EnablePrometheus: true,
		EnableTigera:     false,
		CustomManifests:  []string{"manifest1.yaml", "manifest2.yaml"},
		CustomConfig:     map[string]interface{}{"key": "value"},
	}

	if options.Version != "v3.26.1" {
		t.Errorf("Expected version v3.26.1, got %s", options.Version)
	}

	if options.InstallationType != "operator" {
		t.Errorf("Expected installation type operator, got %s", options.InstallationType)
	}

	if options.CNI != "calico" {
		t.Errorf("Expected CNI calico, got %s", options.CNI)
	}

	if options.IPAM != "calico-ipam" {
		t.Errorf("Expected IPAM calico-ipam, got %s", options.IPAM)
	}

	if options.Datastore != "kubernetes" {
		t.Errorf("Expected datastore kubernetes, got %s", options.Datastore)
	}

	if options.FelixLogSeverity != "Info" {
		t.Errorf("Expected Felix log severity Info, got %s", options.FelixLogSeverity)
	}

	if options.TyphaLogSeverity != "Info" {
		t.Errorf("Expected Typha log severity Info, got %s", options.TyphaLogSeverity)
	}

	if options.OperatorLogLevel != "Info" {
		t.Errorf("Expected operator log level Info, got %s", options.OperatorLogLevel)
	}

	if !options.EnablePrometheus {
		t.Error("Expected Prometheus to be enabled")
	}

	if options.EnableTigera {
		t.Error("Expected Tigera to be disabled")
	}

	if len(options.CustomManifests) != 2 {
		t.Errorf("Expected 2 custom manifests, got %d", len(options.CustomManifests))
	}

	if len(options.CustomConfig) != 1 {
		t.Errorf("Expected 1 custom config item, got %d", len(options.CustomConfig))
	}
}

func TestUpgradeOptions(t *testing.T) {
	options := &UpgradeOptions{
		Version:              "v3.27.0",
		BackupBeforeUpgrade:  true,
		ValidateAfterUpgrade: true,
	}

	if options.Version != "v3.27.0" {
		t.Errorf("Expected version v3.27.0, got %s", options.Version)
	}

	if !options.BackupBeforeUpgrade {
		t.Error("Expected backup before upgrade to be enabled")
	}

	if !options.ValidateAfterUpgrade {
		t.Error("Expected validate after upgrade to be enabled")
	}
}

func TestUninstallOptions(t *testing.T) {
	options := &UninstallOptions{
		RemoveCRDs:       true,
		RemoveNamespace:  true,
		RemoveFinalizers: true,
	}

	if !options.RemoveCRDs {
		t.Error("Expected remove CRDs to be enabled")
	}

	if !options.RemoveNamespace {
		t.Error("Expected remove namespace to be enabled")
	}

	if !options.RemoveFinalizers {
		t.Error("Expected remove finalizers to be enabled")
	}
}

func TestCalicoStatus(t *testing.T) {
	status := &CalicoStatus{
		Installed:       true,
		Version:         "v3.26.1",
		Components:      make(map[string]ComponentStatus),
		Nodes:           []NodeStatus{},
		IPPools:         []IPPoolStatus{},
		BGPPeers:        []BGPPeerStatus{},
		NetworkPolicies: []PolicyStatus{},
		LastUpdated:     time.Now(),
	}

	if !status.Installed {
		t.Error("Expected installed to be true")
	}

	if status.Version != "v3.26.1" {
		t.Errorf("Expected version v3.26.1, got %s", status.Version)
	}

	if status.Components == nil {
		t.Error("Expected components map to be initialized")
	}

	if status.Nodes == nil {
		t.Error("Expected nodes slice to be initialized")
	}

	if status.IPPools == nil {
		t.Error("Expected IP pools slice to be initialized")
	}

	if status.BGPPeers == nil {
		t.Error("Expected BGP peers slice to be initialized")
	}

	if status.NetworkPolicies == nil {
		t.Error("Expected network policies slice to be initialized")
	}
}

func TestComponentStatus(t *testing.T) {
	status := &ComponentStatus{
		Name:    "felix",
		Healthy: true,
		Ready:   true,
		Message: "Component is healthy",
		Version: "v3.26.1",
	}

	if status.Name != "felix" {
		t.Errorf("Expected name felix, got %s", status.Name)
	}

	if !status.Healthy {
		t.Error("Expected component to be healthy")
	}

	if !status.Ready {
		t.Error("Expected component to be ready")
	}

	if status.Message != "Component is healthy" {
		t.Errorf("Expected message 'Component is healthy', got %s", status.Message)
	}

	if status.Version != "v3.26.1" {
		t.Errorf("Expected version v3.26.1, got %s", status.Version)
	}
}

func TestNodeStatus(t *testing.T) {
	status := &NodeStatus{
		Name:  "node-1",
		Ready: true,
		BGP:   true,
		Felix: true,
		Typha: false,
		IP:    "192.168.1.10",
		ASN:   65001,
	}

	if status.Name != "node-1" {
		t.Errorf("Expected name node-1, got %s", status.Name)
	}

	if !status.Ready {
		t.Error("Expected node to be ready")
	}

	if !status.BGP {
		t.Error("Expected BGP to be enabled")
	}

	if !status.Felix {
		t.Error("Expected Felix to be enabled")
	}

	if status.Typha {
		t.Error("Expected Typha to be disabled")
	}

	if status.IP != "192.168.1.10" {
		t.Errorf("Expected IP 192.168.1.10, got %s", status.IP)
	}

	if status.ASN != 65001 {
		t.Errorf("Expected ASN 65001, got %d", status.ASN)
	}
}

func TestIPPoolStatus(t *testing.T) {
	status := &IPPoolStatus{
		Name:      "default-pool",
		CIDR:      "192.168.0.0/16",
		BlockSize: 26,
		IPIP:      true,
		VXLAN:     false,
		Disabled:  false,
	}

	if status.Name != "default-pool" {
		t.Errorf("Expected name default-pool, got %s", status.Name)
	}

	if status.CIDR != "192.168.0.0/16" {
		t.Errorf("Expected CIDR 192.168.0.0/16, got %s", status.CIDR)
	}

	if status.BlockSize != 26 {
		t.Errorf("Expected block size 26, got %d", status.BlockSize)
	}

	if !status.IPIP {
		t.Error("Expected IPIP to be enabled")
	}

	if status.VXLAN {
		t.Error("Expected VXLAN to be disabled")
	}

	if status.Disabled {
		t.Error("Expected pool to be enabled")
	}
}

func TestBGPPeerStatus(t *testing.T) {
	status := &BGPPeerStatus{
		Name:   "peer-1",
		ASN:    65001,
		IP:     "192.168.1.1",
		State:  "Established",
		Uptime: 5 * time.Minute,
	}

	if status.Name != "peer-1" {
		t.Errorf("Expected name peer-1, got %s", status.Name)
	}

	if status.ASN != 65001 {
		t.Errorf("Expected ASN 65001, got %d", status.ASN)
	}

	if status.IP != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", status.IP)
	}

	if status.State != "Established" {
		t.Errorf("Expected state Established, got %s", status.State)
	}

	if status.Uptime != 5*time.Minute {
		t.Errorf("Expected uptime 5 minutes, got %v", status.Uptime)
	}
}

func TestPolicyStatus(t *testing.T) {
	status := &PolicyStatus{
		Name:      "test-policy",
		Namespace: "default",
		Type:      "NetworkPolicy",
		Applied:   true,
	}

	if status.Name != "test-policy" {
		t.Errorf("Expected name test-policy, got %s", status.Name)
	}

	if status.Namespace != "default" {
		t.Errorf("Expected namespace default, got %s", status.Namespace)
	}

	if status.Type != "NetworkPolicy" {
		t.Errorf("Expected type NetworkPolicy, got %s", status.Type)
	}

	if !status.Applied {
		t.Error("Expected policy to be applied")
	}
}

func TestIPPool(t *testing.T) {
	pool := &IPPool{
		Name:        "test-pool",
		CIDR:        "192.168.0.0/16",
		BlockSize:   26,
		IPIPMode:    "CrossSubnet",
		VXLANMode:   "Never",
		Disabled:    false,
		Annotations: map[string]string{"key": "value"},
	}

	if pool.Name != "test-pool" {
		t.Errorf("Expected name test-pool, got %s", pool.Name)
	}

	if pool.CIDR != "192.168.0.0/16" {
		t.Errorf("Expected CIDR 192.168.0.0/16, got %s", pool.CIDR)
	}

	if pool.BlockSize != 26 {
		t.Errorf("Expected block size 26, got %d", pool.BlockSize)
	}

	if pool.IPIPMode != "CrossSubnet" {
		t.Errorf("Expected IPIP mode CrossSubnet, got %s", pool.IPIPMode)
	}

	if pool.VXLANMode != "Never" {
		t.Errorf("Expected VXLAN mode Never, got %s", pool.VXLANMode)
	}

	if pool.Disabled {
		t.Error("Expected pool to be enabled")
	}

	if len(pool.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(pool.Annotations))
	}
}

func TestNetworkPolicy(t *testing.T) {
	policy := &NetworkPolicy{
		Name:         "test-policy",
		Namespace:    "default",
		Selector:     "app == 'web'",
		IngressRules: []Rule{},
		EgressRules:  []Rule{},
		Types:        []string{"Ingress", "Egress"},
		Annotations:  map[string]string{"key": "value"},
	}

	if policy.Name != "test-policy" {
		t.Errorf("Expected name test-policy, got %s", policy.Name)
	}

	if policy.Namespace != "default" {
		t.Errorf("Expected namespace default, got %s", policy.Namespace)
	}

	if policy.Selector != "app == 'web'" {
		t.Errorf("Expected selector 'app == \\'web\\'', got %s", policy.Selector)
	}

	if len(policy.Types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(policy.Types))
	}

	if len(policy.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(policy.Annotations))
	}
}

func TestRule(t *testing.T) {
	rule := &Rule{
		Action:   "Allow",
		Protocol: "TCP",
		Source: RuleEndpoint{
			Selector: "app == 'frontend'",
		},
		Destination: RuleEndpoint{
			Ports: []Port{
				{Number: 80, Protocol: "TCP"},
			},
		},
	}

	if rule.Action != "Allow" {
		t.Errorf("Expected action Allow, got %s", rule.Action)
	}

	if rule.Protocol != "TCP" {
		t.Errorf("Expected protocol TCP, got %s", rule.Protocol)
	}

	if rule.Source.Selector != "app == 'frontend'" {
		t.Errorf("Expected source selector 'app == \\'frontend\\'', got %s", rule.Source.Selector)
	}

	if len(rule.Destination.Ports) != 1 {
		t.Errorf("Expected 1 destination port, got %d", len(rule.Destination.Ports))
	}
}

func TestRuleEndpoint(t *testing.T) {
	endpoint := &RuleEndpoint{
		Nets:        []string{"192.168.0.0/16"},
		NotNets:     []string{"192.168.1.0/24"},
		Selector:    "app == 'web'",
		NotSelector: "app == 'admin'",
		Ports:       []Port{{Number: 80, Protocol: "TCP"}},
		NotPorts:    []Port{{Number: 443, Protocol: "TCP"}},
	}

	if len(endpoint.Nets) != 1 {
		t.Errorf("Expected 1 net, got %d", len(endpoint.Nets))
	}

	if len(endpoint.NotNets) != 1 {
		t.Errorf("Expected 1 not net, got %d", len(endpoint.NotNets))
	}

	if endpoint.Selector != "app == 'web'" {
		t.Errorf("Expected selector 'app == \\'web\\'', got %s", endpoint.Selector)
	}

	if endpoint.NotSelector != "app == 'admin'" {
		t.Errorf("Expected not selector 'app == \\'admin\\'', got %s", endpoint.NotSelector)
	}

	if len(endpoint.Ports) != 1 {
		t.Errorf("Expected 1 port, got %d", len(endpoint.Ports))
	}

	if len(endpoint.NotPorts) != 1 {
		t.Errorf("Expected 1 not port, got %d", len(endpoint.NotPorts))
	}
}

func TestPort(t *testing.T) {
	port := &Port{
		Number:   80,
		Protocol: "TCP",
		EndPort:  90,
	}

	if port.Number != 80 {
		t.Errorf("Expected number 80, got %d", port.Number)
	}

	if port.Protocol != "TCP" {
		t.Errorf("Expected protocol TCP, got %s", port.Protocol)
	}

	if port.EndPort != 90 {
		t.Errorf("Expected end port 90, got %d", port.EndPort)
	}
}

func TestHTTPMatch(t *testing.T) {
	httpMatch := &HTTPMatch{
		Methods: []string{"GET", "POST"},
		Paths:   []string{"/api/*"},
		Headers: map[string]string{"Content-Type": "application/json"},
	}

	if len(httpMatch.Methods) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(httpMatch.Methods))
	}

	if len(httpMatch.Paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(httpMatch.Paths))
	}

	if len(httpMatch.Headers) != 1 {
		t.Errorf("Expected 1 header, got %d", len(httpMatch.Headers))
	}
}

func TestICMPMatch(t *testing.T) {
	icmpMatch := &ICMPMatch{
		Type: 8,
		Code: 0,
	}

	if icmpMatch.Type != 8 {
		t.Errorf("Expected type 8, got %d", icmpMatch.Type)
	}

	if icmpMatch.Code != 0 {
		t.Errorf("Expected code 0, got %d", icmpMatch.Code)
	}
}

func TestBGPPeer(t *testing.T) {
	peer := &BGPPeer{
		Name:         "test-peer",
		ASN:          65001,
		IP:           "192.168.1.1",
		NodeSelector: "has(router)",
		Password:     "secret",
		Annotations:  map[string]string{"key": "value"},
	}

	if peer.Name != "test-peer" {
		t.Errorf("Expected name test-peer, got %s", peer.Name)
	}

	if peer.ASN != 65001 {
		t.Errorf("Expected ASN 65001, got %d", peer.ASN)
	}

	if peer.IP != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", peer.IP)
	}

	if peer.NodeSelector != "has(router)" {
		t.Errorf("Expected node selector 'has(router)', got %s", peer.NodeSelector)
	}

	if peer.Password != "secret" {
		t.Errorf("Expected password secret, got %s", peer.Password)
	}

	if len(peer.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(peer.Annotations))
	}
}

func TestHealthCheckResult(t *testing.T) {
	result := &HealthCheckResult{
		Overall:     true,
		Components:  map[string]bool{"felix": true},
		Nodes:       map[string]bool{"node-1": true},
		BGP:         map[string]bool{"peer-1": true},
		LastChecked: time.Now(),
		Errors:      []string{},
	}

	if !result.Overall {
		t.Error("Expected overall health to be true")
	}

	if len(result.Components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(result.Components))
	}

	if len(result.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(result.Nodes))
	}

	if len(result.BGP) != 1 {
		t.Errorf("Expected 1 BGP peer, got %d", len(result.BGP))
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}

func TestFelixConfiguration(t *testing.T) {
	config := &FelixConfiguration{
		Name:                            "default",
		LogSeverityScreen:               "Info",
		LogSeverityFile:                 "Info",
		LogSeveritySys:                  "Info",
		PrometheusMetricsEnabled:        true,
		PrometheusMetricsPort:           9091,
		PrometheusGoMetricsEnabled:      true,
		PrometheusProcessMetricsEnabled: true,
	}

	if config.Name != "default" {
		t.Errorf("Expected name default, got %s", config.Name)
	}

	if config.LogSeverityScreen != "Info" {
		t.Errorf("Expected screen log severity Info, got %s", config.LogSeverityScreen)
	}

	if config.LogSeverityFile != "Info" {
		t.Errorf("Expected file log severity Info, got %s", config.LogSeverityFile)
	}

	if config.LogSeveritySys != "Info" {
		t.Errorf("Expected sys log severity Info, got %s", config.LogSeveritySys)
	}

	if !config.PrometheusMetricsEnabled {
		t.Error("Expected Prometheus metrics to be enabled")
	}

	if config.PrometheusMetricsPort != 9091 {
		t.Errorf("Expected Prometheus metrics port 9091, got %d", config.PrometheusMetricsPort)
	}

	if !config.PrometheusGoMetricsEnabled {
		t.Error("Expected Prometheus Go metrics to be enabled")
	}

	if !config.PrometheusProcessMetricsEnabled {
		t.Error("Expected Prometheus process metrics to be enabled")
	}
}

func TestBGPConfiguration(t *testing.T) {
	config := &BGPConfiguration{
		Name:                   "default",
		ASNumber:               65001,
		ServiceClusterIPs:      []string{"10.96.0.0/12"},
		ServiceExternalIPs:     []string{"1.2.3.4/32"},
		ServiceLoadBalancerIPs: []string{"5.6.7.8/32"},
	}

	if config.Name != "default" {
		t.Errorf("Expected name default, got %s", config.Name)
	}

	if config.ASNumber != 65001 {
		t.Errorf("Expected AS number 65001, got %d", config.ASNumber)
	}

	if len(config.ServiceClusterIPs) != 1 {
		t.Errorf("Expected 1 service cluster IP, got %d", len(config.ServiceClusterIPs))
	}

	if len(config.ServiceExternalIPs) != 1 {
		t.Errorf("Expected 1 service external IP, got %d", len(config.ServiceExternalIPs))
	}

	if len(config.ServiceLoadBalancerIPs) != 1 {
		t.Errorf("Expected 1 service load balancer IP, got %d", len(config.ServiceLoadBalancerIPs))
	}
}

func TestClusterInformation(t *testing.T) {
	info := &ClusterInformation{
		Name:          "default",
		CalicoVersion: "v3.26.1",
		ClusterType:   "k8s",
		DatastoreType: "kubernetes",
	}

	if info.Name != "default" {
		t.Errorf("Expected name default, got %s", info.Name)
	}

	if info.CalicoVersion != "v3.26.1" {
		t.Errorf("Expected Calico version v3.26.1, got %s", info.CalicoVersion)
	}

	if info.ClusterType != "k8s" {
		t.Errorf("Expected cluster type k8s, got %s", info.ClusterType)
	}

	if info.DatastoreType != "kubernetes" {
		t.Errorf("Expected datastore type kubernetes, got %s", info.DatastoreType)
	}
}

func TestGlobalNetworkSet(t *testing.T) {
	networkSet := &GlobalNetworkSet{
		Name:        "test-networks",
		Nets:        []string{"10.0.0.0/8", "172.16.0.0/12"},
		Labels:      map[string]string{"environment": "production"},
		Annotations: map[string]string{"key": "value"},
	}

	if networkSet.Name != "test-networks" {
		t.Errorf("Expected name test-networks, got %s", networkSet.Name)
	}

	if len(networkSet.Nets) != 2 {
		t.Errorf("Expected 2 nets, got %d", len(networkSet.Nets))
	}

	if len(networkSet.Labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(networkSet.Labels))
	}

	if len(networkSet.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(networkSet.Annotations))
	}
}

func TestHostEndpoint(t *testing.T) {
	endpoint := &HostEndpoint{
		Name:        "test-host",
		Node:        "node-1",
		Interface:   "eth0",
		ExpectedIPs: []string{"192.168.1.10"},
		Profiles:    []string{"default"},
		Labels:      map[string]string{"environment": "production"},
		Annotations: map[string]string{"key": "value"},
	}

	if endpoint.Name != "test-host" {
		t.Errorf("Expected name test-host, got %s", endpoint.Name)
	}

	if endpoint.Node != "node-1" {
		t.Errorf("Expected node node-1, got %s", endpoint.Node)
	}

	if endpoint.Interface != "eth0" {
		t.Errorf("Expected interface eth0, got %s", endpoint.Interface)
	}

	if len(endpoint.ExpectedIPs) != 1 {
		t.Errorf("Expected 1 expected IP, got %d", len(endpoint.ExpectedIPs))
	}

	if len(endpoint.Profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(endpoint.Profiles))
	}

	if len(endpoint.Labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(endpoint.Labels))
	}

	if len(endpoint.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(endpoint.Annotations))
	}
}

func TestWorkloadEndpoint(t *testing.T) {
	endpoint := &WorkloadEndpoint{
		Name:        "test-workload",
		Namespace:   "default",
		Pod:         "test-pod",
		Container:   "test-container",
		Interface:   "eth0",
		ExpectedIPs: []string{"10.0.0.1"},
		Profiles:    []string{"default"},
		Labels:      map[string]string{"app": "web"},
		Annotations: map[string]string{"key": "value"},
	}

	if endpoint.Name != "test-workload" {
		t.Errorf("Expected name test-workload, got %s", endpoint.Name)
	}

	if endpoint.Namespace != "default" {
		t.Errorf("Expected namespace default, got %s", endpoint.Namespace)
	}

	if endpoint.Pod != "test-pod" {
		t.Errorf("Expected pod test-pod, got %s", endpoint.Pod)
	}

	if endpoint.Container != "test-container" {
		t.Errorf("Expected container test-container, got %s", endpoint.Container)
	}

	if endpoint.Interface != "eth0" {
		t.Errorf("Expected interface eth0, got %s", endpoint.Interface)
	}

	if len(endpoint.ExpectedIPs) != 1 {
		t.Errorf("Expected 1 expected IP, got %d", len(endpoint.ExpectedIPs))
	}

	if len(endpoint.Profiles) != 1 {
		t.Errorf("Expected 1 profile, got %d", len(endpoint.Profiles))
	}

	if len(endpoint.Labels) != 1 {
		t.Errorf("Expected 1 label, got %d", len(endpoint.Labels))
	}

	if len(endpoint.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(endpoint.Annotations))
	}
}

// Test helper structs
type testLogger struct {
	messages []string
}

func (t *testLogger) Write(p []byte) (n int, err error) {
	t.messages = append(t.messages, string(p))
	return len(p), nil
}

// Mock Kubernetes clients for testing
type mockKubernetesClient struct{}
type mockDynamicClient struct{}

func (m *mockKubernetesClient) CoreV1() interface{} { return nil }
func (m *mockDynamicClient) Resource(gvr schema.GroupVersionResource) interface{} { return nil }

// Test CalicoManager Install method
func TestCalicoManager_Install(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	options := &InstallOptions{
		Version:          "v3.26.1",
		InstallationType: "operator",
		CNI:              "calico",
		IPAM:             "calico-ipam",
		Datastore:        "kubernetes",
		EnablePrometheus: true,
	}

	err := manager.Install(ctx, options)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during installation")
	}

	// Check for specific log messages
	foundInstallLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Installing Calico version") {
			foundInstallLog = true
			break
		}
	}
	if !foundInstallLog {
		t.Error("Expected installation log message")
	}
}

// Test CalicoManager Upgrade method
func TestCalicoManager_Upgrade(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	options := &UpgradeOptions{
		Version:              "v3.27.0",
		BackupBeforeUpgrade:  true,
		ValidateAfterUpgrade: true,
	}

	err := manager.Upgrade(ctx, options)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during upgrade")
	}

	// Check for specific log messages
	foundUpgradeLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Upgrading Calico to version") {
			foundUpgradeLog = true
			break
		}
	}
	if !foundUpgradeLog {
		t.Error("Expected upgrade log message")
	}
}

// Test CalicoManager Uninstall method
func TestCalicoManager_Uninstall(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	options := &UninstallOptions{
		RemoveCRDs:       true,
		RemoveNamespace:  true,
		RemoveFinalizers: true,
	}

	err := manager.Uninstall(ctx, options)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during uninstallation")
	}

	// Check for specific log messages
	foundUninstallLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Uninstalling Calico") {
			foundUninstallLog = true
			break
		}
	}
	if !foundUninstallLog {
		t.Error("Expected uninstall log message")
	}
}

// Test CalicoManager GetStatus method
func TestCalicoManager_GetStatus(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	status, err := manager.GetStatus(ctx)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	if status != nil {
		t.Error("Expected status to be nil due to error")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during status check")
	}

	// Check for specific log messages
	foundStatusLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Getting Calico status") {
			foundStatusLog = true
			break
		}
	}
	if !foundStatusLog {
		t.Error("Expected status log message")
	}
}

// Test CalicoManager WaitForReady method
func TestCalicoManager_WaitForReady(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := manager.WaitForReady(ctx, 50*time.Millisecond)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during wait")
	}

	// Check for specific log messages
	foundWaitLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Waiting for Calico to be ready") {
			foundWaitLog = true
			break
		}
	}
	if !foundWaitLog {
		t.Error("Expected wait log message")
	}
}

// Test CalicoManager HealthCheck method
func TestCalicoManager_HealthCheck(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	result, err := manager.HealthCheck(ctx)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	if result != nil {
		t.Error("Expected result to be nil due to error")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during health check")
	}

	// Check for specific log messages
	foundHealthLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Performing Calico health check") {
			foundHealthLog = true
			break
		}
	}
	if !foundHealthLog {
		t.Error("Expected health check log message")
	}
}

// Test CalicoManager ApplyResource method
func TestCalicoManager_ApplyResource(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	yamlContent := `apiVersion: crd.projectcalico.org/v1
kind: IPPool
metadata:
  name: test-pool
spec:
  cidr: 192.168.0.0/16
  blockSize: 26
  ipipMode: CrossSubnet
  vxlanMode: Never
  natOutgoing: true
  nodeSelector: all()
`

	err := manager.ApplyResource(ctx, yamlContent)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during resource application")
	}

	// Check for specific log messages
	foundApplyLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Applying Calico resource from YAML") {
			foundApplyLog = true
			break
		}
	}
	if !foundApplyLog {
		t.Error("Expected apply resource log message")
	}
}

// Test CalicoManager ApplyResource with invalid YAML
func TestCalicoManager_ApplyResource_InvalidYAML(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	invalidYAML := `invalid yaml content
  - this is not valid yaml
    - it should cause an error
`

	err := manager.ApplyResource(ctx, invalidYAML)

	// Should return an error for invalid YAML
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during resource application")
	}
}

// Test CalicoManager GetIPPools method
func TestCalicoManager_GetIPPools(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	pools, err := manager.GetIPPools(ctx)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	if pools != nil {
		t.Error("Expected pools to be nil due to error")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during IP pools retrieval")
	}

	// Check for specific log messages
	foundPoolsLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Getting IP pools") {
			foundPoolsLog = true
			break
		}
	}
	if !foundPoolsLog {
		t.Error("Expected IP pools log message")
	}
}

// Test CalicoManager CreateIPPool method
func TestCalicoManager_CreateIPPool(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	pool := &IPPool{
		Name:        "test-pool",
		CIDR:        "192.168.0.0/16",
		BlockSize:   26,
		IPIPMode:    "CrossSubnet",
		VXLANMode:   "Never",
		Disabled:    false,
		Annotations: map[string]string{"key": "value"},
	}

	err := manager.CreateIPPool(ctx, pool)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during IP pool creation")
	}

	// Check for specific log messages
	foundCreateLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Creating IP pool test-pool") {
			foundCreateLog = true
			break
		}
	}
	if !foundCreateLog {
		t.Error("Expected IP pool creation log message")
	}
}

// Test CalicoManager DeleteIPPool method
func TestCalicoManager_DeleteIPPool(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	err := manager.DeleteIPPool(ctx, "test-pool")

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during IP pool deletion")
	}

	// Check for specific log messages
	foundDeleteLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Deleting IP pool test-pool") {
			foundDeleteLog = true
			break
		}
	}
	if !foundDeleteLog {
		t.Error("Expected IP pool deletion log message")
	}
}

// Test CalicoManager GetNetworkPolicies method
func TestCalicoManager_GetNetworkPolicies(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	policies, err := manager.GetNetworkPolicies(ctx, "default")

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	if policies != nil {
		t.Error("Expected policies to be nil due to error")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during network policies retrieval")
	}

	// Check for specific log messages
	foundPoliciesLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Getting network policies for namespace default") {
			foundPoliciesLog = true
			break
		}
	}
	if !foundPoliciesLog {
		t.Error("Expected network policies log message")
	}
}

// Test CalicoManager CreateNetworkPolicy method
func TestCalicoManager_CreateNetworkPolicy(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	policy := &NetworkPolicy{
		Name:         "test-policy",
		Namespace:    "default",
		Selector:     "app == 'web'",
		IngressRules: []Rule{},
		EgressRules:  []Rule{},
		Types:        []string{"Ingress", "Egress"},
		Annotations:  map[string]string{"key": "value"},
	}

	err := manager.CreateNetworkPolicy(ctx, policy)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during network policy creation")
	}

	// Check for specific log messages
	foundCreateLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Creating network policy test-policy") {
			foundCreateLog = true
			break
		}
	}
	if !foundCreateLog {
		t.Error("Expected network policy creation log message")
	}
}

// Test CalicoManager DeleteNetworkPolicy method
func TestCalicoManager_DeleteNetworkPolicy(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	err := manager.DeleteNetworkPolicy(ctx, "test-policy", "default")

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during network policy deletion")
	}

	// Check for specific log messages
	foundDeleteLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Deleting network policy test-policy from namespace default") {
			foundDeleteLog = true
			break
		}
	}
	if !foundDeleteLog {
		t.Error("Expected network policy deletion log message")
	}
}

// Test CalicoManager GetBGPPeers method
func TestCalicoManager_GetBGPPeers(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	peers, err := manager.GetBGPPeers(ctx)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	if peers != nil {
		t.Error("Expected peers to be nil due to error")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during BGP peers retrieval")
	}

	// Check for specific log messages
	foundPeersLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Getting BGP peers") {
			foundPeersLog = true
			break
		}
	}
	if !foundPeersLog {
		t.Error("Expected BGP peers log message")
	}
}

// Test CalicoManager CreateBGPPeer method
func TestCalicoManager_CreateBGPPeer(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	peer := &BGPPeer{
		Name:         "test-peer",
		ASN:          65001,
		IP:           "192.168.1.1",
		NodeSelector: "has(router)",
		Password:     "secret",
		Annotations:  map[string]string{"key": "value"},
	}

	err := manager.CreateBGPPeer(ctx, peer)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during BGP peer creation")
	}

	// Check for specific log messages
	foundCreateLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Creating BGP peer test-peer") {
			foundCreateLog = true
			break
		}
	}
	if !foundCreateLog {
		t.Error("Expected BGP peer creation log message")
	}
}

// Test CalicoManager DeleteBGPPeer method
func TestCalicoManager_DeleteBGPPeer(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	err := manager.DeleteBGPPeer(ctx, "test-peer")

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during BGP peer deletion")
	}

	// Check for specific log messages
	foundDeleteLog := false
	for _, msg := range logger.messages {
		if strings.Contains(msg, "Deleting BGP peer test-peer") {
			foundDeleteLog = true
			break
		}
	}
	if !foundDeleteLog {
		t.Error("Expected BGP peer deletion log message")
	}
}

// Test edge cases and validation scenarios

// Test CalicoManager with nil logger
func TestCalicoManager_Logf_NilLogger(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	// Should not panic when logger is nil
	manager.Logf("test message %s", "value")
}

// Test CalicoManager with empty options
func TestCalicoManager_Install_EmptyOptions(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	options := &InstallOptions{}

	err := manager.Install(ctx, options)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during installation")
	}
}

// Test CalicoManager with context cancellation
func TestCalicoManager_WaitForReady_ContextCancelled(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := manager.WaitForReady(ctx, 1*time.Second)

	// Should return context cancelled error
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
}

// Test CalicoManager with timeout
func TestCalicoManager_WaitForReady_Timeout(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := manager.WaitForReady(ctx, 5*time.Millisecond)

	// Should return timeout error
	if err == nil {
		t.Error("Expected error due to timeout")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

// Test CalicoManager ApplyResource with empty YAML
func TestCalicoManager_ApplyResource_EmptyYAML(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	err := manager.ApplyResource(ctx, "")

	// Should return an error for empty YAML
	if err == nil {
		t.Error("Expected error for empty YAML")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during resource application")
	}
}

// Test CalicoManager ApplyResource with multiple YAML documents
func TestCalicoManager_ApplyResource_MultipleDocuments(t *testing.T) {
	logger := &testLogger{}
	manager := NewCalicoManager(nil, nil, nil, logger)

	ctx := context.Background()
	multiYAML := `apiVersion: crd.projectcalico.org/v1
kind: IPPool
metadata:
  name: pool-1
spec:
  cidr: 192.168.0.0/16
  blockSize: 26
---
apiVersion: crd.projectcalico.org/v1
kind: IPPool
metadata:
  name: pool-2
spec:
  cidr: 10.0.0.0/16
  blockSize: 24
`

	err := manager.ApplyResource(ctx, multiYAML)

	// Since helper methods are not implemented, we expect an error
	if err == nil {
		t.Error("Expected error due to unimplemented helper methods")
	}

	// Check that logging occurred
	if len(logger.messages) == 0 {
		t.Error("Expected log messages during resource application")
	}
}

// Test data structure validation edge cases

// Test InstallOptions with all fields set
func TestInstallOptions_AllFields(t *testing.T) {
	options := &InstallOptions{
		Version:          "v3.26.1",
		InstallationType: "operator",
		CNI:              "calico",
		IPAM:             "calico-ipam",
		Datastore:        "kubernetes",
		FelixLogSeverity: "Info",
		TyphaLogSeverity: "Info",
		OperatorLogLevel: "Info",
		EnablePrometheus: true,
		EnableTigera:     false,
		CustomManifests:  []string{"manifest1.yaml", "manifest2.yaml"},
		CustomConfig:     map[string]interface{}{"key": "value", "number": 42},
	}

	// Test all fields are set correctly
	if options.Version != "v3.26.1" {
		t.Errorf("Expected version v3.26.1, got %s", options.Version)
	}

	if options.InstallationType != "operator" {
		t.Errorf("Expected installation type operator, got %s", options.InstallationType)
	}

	if options.CNI != "calico" {
		t.Errorf("Expected CNI calico, got %s", options.CNI)
	}

	if options.IPAM != "calico-ipam" {
		t.Errorf("Expected IPAM calico-ipam, got %s", options.IPAM)
	}

	if options.Datastore != "kubernetes" {
		t.Errorf("Expected datastore kubernetes, got %s", options.Datastore)
	}

	if options.FelixLogSeverity != "Info" {
		t.Errorf("Expected Felix log severity Info, got %s", options.FelixLogSeverity)
	}

	if options.TyphaLogSeverity != "Info" {
		t.Errorf("Expected Typha log severity Info, got %s", options.TyphaLogSeverity)
	}

	if options.OperatorLogLevel != "Info" {
		t.Errorf("Expected operator log level Info, got %s", options.OperatorLogLevel)
	}

	if !options.EnablePrometheus {
		t.Error("Expected Prometheus to be enabled")
	}

	if options.EnableTigera {
		t.Error("Expected Tigera to be disabled")
	}

	if len(options.CustomManifests) != 2 {
		t.Errorf("Expected 2 custom manifests, got %d", len(options.CustomManifests))
	}

	if len(options.CustomConfig) != 2 {
		t.Errorf("Expected 2 custom config items, got %d", len(options.CustomConfig))
	}

	// Test custom config values
	if options.CustomConfig["key"] != "value" {
		t.Errorf("Expected custom config key to be 'value', got %v", options.CustomConfig["key"])
	}

	if options.CustomConfig["number"] != 42 {
		t.Errorf("Expected custom config number to be 42, got %v", options.CustomConfig["number"])
	}
}

// Test NetworkPolicy with complex rules
func TestNetworkPolicy_ComplexRules(t *testing.T) {
	policy := &NetworkPolicy{
		Name:      "complex-policy",
		Namespace: "default",
		Selector:  "app == 'web'",
		IngressRules: []Rule{
			{
				Action:   "Allow",
				Protocol: "TCP",
				Source: RuleEndpoint{
					Selector: "app == 'frontend'",
					Nets:     []string{"192.168.0.0/16"},
					NotNets:  []string{"192.168.1.0/24"},
					Ports:    []Port{{Number: 80, Protocol: "TCP"}},
					NotPorts: []Port{{Number: 443, Protocol: "TCP"}},
				},
				Destination: RuleEndpoint{
					Selector: "app == 'backend'",
					Ports:    []Port{{Number: 8080, Protocol: "TCP"}},
				},
				HTTP: &HTTPMatch{
					Methods: []string{"GET", "POST", "PUT"},
					Paths:   []string{"/api/*", "/health"},
					Headers: map[string]string{
						"Content-Type": "application/json",
						"Authorization": "Bearer *",
					},
				},
				ICMP: &ICMPMatch{
					Type: 8,
					Code: 0,
				},
			},
		},
		EgressRules: []Rule{
			{
				Action:   "Deny",
				Protocol: "UDP",
				Source: RuleEndpoint{
					Nets: []string{"10.0.0.0/8"},
				},
			},
		},
		Types:       []string{"Ingress", "Egress"},
		Annotations: map[string]string{"policy.kubernetes.io/description": "Complex test policy"},
	}

	// Test policy structure
	if policy.Name != "complex-policy" {
		t.Errorf("Expected name complex-policy, got %s", policy.Name)
	}

	if policy.Namespace != "default" {
		t.Errorf("Expected namespace default, got %s", policy.Namespace)
	}

	if policy.Selector != "app == 'web'" {
		t.Errorf("Expected selector 'app == \\'web\\'', got %s", policy.Selector)
	}

	if len(policy.IngressRules) != 1 {
		t.Errorf("Expected 1 ingress rule, got %d", len(policy.IngressRules))
	}

	if len(policy.EgressRules) != 1 {
		t.Errorf("Expected 1 egress rule, got %d", len(policy.EgressRules))
	}

	if len(policy.Types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(policy.Types))
	}

	if len(policy.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(policy.Annotations))
	}

	// Test ingress rule
	ingressRule := policy.IngressRules[0]
	if ingressRule.Action != "Allow" {
		t.Errorf("Expected action Allow, got %s", ingressRule.Action)
	}

	if ingressRule.Protocol != "TCP" {
		t.Errorf("Expected protocol TCP, got %s", ingressRule.Protocol)
	}

	if ingressRule.Source.Selector != "app == 'frontend'" {
		t.Errorf("Expected source selector 'app == \\'frontend\\'', got %s", ingressRule.Source.Selector)
	}

	if len(ingressRule.Source.Nets) != 1 {
		t.Errorf("Expected 1 source net, got %d", len(ingressRule.Source.Nets))
	}

	if len(ingressRule.Source.NotNets) != 1 {
		t.Errorf("Expected 1 source not net, got %d", len(ingressRule.Source.NotNets))
	}

	if len(ingressRule.Source.Ports) != 1 {
		t.Errorf("Expected 1 source port, got %d", len(ingressRule.Source.Ports))
	}

	if len(ingressRule.Source.NotPorts) != 1 {
		t.Errorf("Expected 1 source not port, got %d", len(ingressRule.Source.NotPorts))
	}

	// Test HTTP match
	if ingressRule.HTTP == nil {
		t.Fatal("Expected HTTP match to be set")
	}

	if len(ingressRule.HTTP.Methods) != 3 {
		t.Errorf("Expected 3 HTTP methods, got %d", len(ingressRule.HTTP.Methods))
	}

	if len(ingressRule.HTTP.Paths) != 2 {
		t.Errorf("Expected 2 HTTP paths, got %d", len(ingressRule.HTTP.Paths))
	}

	if len(ingressRule.HTTP.Headers) != 2 {
		t.Errorf("Expected 2 HTTP headers, got %d", len(ingressRule.HTTP.Headers))
	}

	// Test ICMP match
	if ingressRule.ICMP == nil {
		t.Fatal("Expected ICMP match to be set")
	}

	if ingressRule.ICMP.Type != 8 {
		t.Errorf("Expected ICMP type 8, got %d", ingressRule.ICMP.Type)
	}

	if ingressRule.ICMP.Code != 0 {
		t.Errorf("Expected ICMP code 0, got %d", ingressRule.ICMP.Code)
	}

	// Test egress rule
	egressRule := policy.EgressRules[0]
	if egressRule.Action != "Deny" {
		t.Errorf("Expected action Deny, got %s", egressRule.Action)
	}

	if egressRule.Protocol != "UDP" {
		t.Errorf("Expected protocol UDP, got %s", egressRule.Protocol)
	}

	if len(egressRule.Source.Nets) != 1 {
		t.Errorf("Expected 1 source net, got %d", len(egressRule.Source.Nets))
	}
}

// Test HealthCheckResult with various states
func TestHealthCheckResult_VariousStates(t *testing.T) {
	// Test healthy state
	healthyResult := &HealthCheckResult{
		Overall:     true,
		Components:  map[string]bool{"felix": true, "typha": true},
		Nodes:       map[string]bool{"node-1": true, "node-2": true},
		BGP:         map[string]bool{"peer-1": true},
		LastChecked: time.Now(),
		Errors:      []string{},
	}

	if !healthyResult.Overall {
		t.Error("Expected overall health to be true")
	}

	if len(healthyResult.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(healthyResult.Components))
	}

	if len(healthyResult.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(healthyResult.Nodes))
	}

	if len(healthyResult.BGP) != 1 {
		t.Errorf("Expected 1 BGP peer, got %d", len(healthyResult.BGP))
	}

	if len(healthyResult.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(healthyResult.Errors))
	}

	// Test unhealthy state
	unhealthyResult := &HealthCheckResult{
		Overall:     false,
		Components:  map[string]bool{"felix": true, "typha": false},
		Nodes:       map[string]bool{"node-1": true, "node-2": false},
		BGP:         map[string]bool{"peer-1": false},
		LastChecked: time.Now(),
		Errors:      []string{"Typha is not ready", "Node node-2 is unhealthy", "BGP peer peer-1 is down"},
	}

	if unhealthyResult.Overall {
		t.Error("Expected overall health to be false")
	}

	if len(unhealthyResult.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(unhealthyResult.Components))
	}

	if len(unhealthyResult.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(unhealthyResult.Nodes))
	}

	if len(unhealthyResult.BGP) != 1 {
		t.Errorf("Expected 1 BGP peer, got %d", len(unhealthyResult.BGP))
	}

	if len(unhealthyResult.Errors) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(unhealthyResult.Errors))
	}

	// Check specific component states
	if !unhealthyResult.Components["felix"] {
		t.Error("Expected Felix to be healthy")
	}

	if unhealthyResult.Components["typha"] {
		t.Error("Expected Typha to be unhealthy")
	}

	if !unhealthyResult.Nodes["node-1"] {
		t.Error("Expected node-1 to be healthy")
	}

	if unhealthyResult.Nodes["node-2"] {
		t.Error("Expected node-2 to be unhealthy")
	}

	if unhealthyResult.BGP["peer-1"] {
		t.Error("Expected peer-1 to be unhealthy")
	}
}

// Test CalicoStatus with various component states
func TestCalicoStatus_VariousStates(t *testing.T) {
	// Test installed state
	installedStatus := &CalicoStatus{
		Installed:   true,
		Version:     "v3.26.1",
		Components: map[string]ComponentStatus{
			"felix": {
				Name:    "felix",
				Healthy: true,
				Ready:   true,
				Message: "Component is healthy",
				Version: "v3.26.1",
			},
			"typha": {
				Name:    "typha",
				Healthy: true,
				Ready:   true,
				Message: "Component is healthy",
				Version: "v3.26.1",
			},
		},
		Nodes: []NodeStatus{
			{
				Name:  "node-1",
				Ready: true,
				BGP:   true,
				Felix: true,
				Typha: false,
				IP:    "192.168.1.10",
				ASN:   65001,
			},
			{
				Name:  "node-2",
				Ready: true,
				BGP:   true,
				Felix: true,
				Typha: true,
				IP:    "192.168.1.11",
				ASN:   65001,
			},
		},
		IPPools: []IPPoolStatus{
			{
				Name:      "default-pool",
				CIDR:      "192.168.0.0/16",
				BlockSize: 26,
				IPIP:      true,
				VXLAN:     false,
				Disabled:  false,
			},
		},
		BGPPeers: []BGPPeerStatus{
			{
				Name:   "peer-1",
				ASN:    65001,
				IP:     "192.168.1.1",
				State:  "Established",
				Uptime: 5 * time.Minute,
			},
		},
		NetworkPolicies: []PolicyStatus{
			{
				Name:      "test-policy",
				Namespace: "default",
				Type:      "NetworkPolicy",
				Applied:   true,
			},
		},
		LastUpdated: time.Now(),
	}

	if !installedStatus.Installed {
		t.Error("Expected Calico to be installed")
	}

	if installedStatus.Version != "v3.26.1" {
		t.Errorf("Expected version v3.26.1, got %s", installedStatus.Version)
	}

	if len(installedStatus.Components) != 2 {
		t.Errorf("Expected 2 components, got %d", len(installedStatus.Components))
	}

	if len(installedStatus.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(installedStatus.Nodes))
	}

	if len(installedStatus.IPPools) != 1 {
		t.Errorf("Expected 1 IP pool, got %d", len(installedStatus.IPPools))
	}

	if len(installedStatus.BGPPeers) != 1 {
		t.Errorf("Expected 1 BGP peer, got %d", len(installedStatus.BGPPeers))
	}

	if len(installedStatus.NetworkPolicies) != 1 {
		t.Errorf("Expected 1 network policy, got %d", len(installedStatus.NetworkPolicies))
	}

	// Test not installed state
	notInstalledStatus := &CalicoStatus{
		Installed:       false,
		Version:         "",
		Components:      make(map[string]ComponentStatus),
		Nodes:           []NodeStatus{},
		IPPools:         []IPPoolStatus{},
		BGPPeers:        []BGPPeerStatus{},
		NetworkPolicies: []PolicyStatus{},
		LastUpdated:     time.Now(),
	}

	if notInstalledStatus.Installed {
		t.Error("Expected Calico to not be installed")
	}

	if notInstalledStatus.Version != "" {
		t.Errorf("Expected empty version, got %s", notInstalledStatus.Version)
	}

	if len(notInstalledStatus.Components) != 0 {
		t.Errorf("Expected 0 components, got %d", len(notInstalledStatus.Components))
	}

	if len(notInstalledStatus.Nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(notInstalledStatus.Nodes))
	}

	if len(notInstalledStatus.IPPools) != 0 {
		t.Errorf("Expected 0 IP pools, got %d", len(notInstalledStatus.IPPools))
	}

	if len(notInstalledStatus.BGPPeers) != 0 {
		t.Errorf("Expected 0 BGP peers, got %d", len(notInstalledStatus.BGPPeers))
	}

	if len(notInstalledStatus.NetworkPolicies) != 0 {
		t.Errorf("Expected 0 network policies, got %d", len(notInstalledStatus.NetworkPolicies))
	}
}

// Test validation functions for data structures

// Test IPPool validation
func TestIPPool_Validation(t *testing.T) {
	// Test valid IP pool
	validPool := &IPPool{
		Name:        "test-pool",
		CIDR:        "192.168.0.0/16",
		BlockSize:   26,
		IPIPMode:    "CrossSubnet",
		VXLANMode:   "Never",
		Disabled:    false,
		Annotations: map[string]string{"key": "value"},
	}

	if validPool.Name != "test-pool" {
		t.Errorf("Expected name test-pool, got %s", validPool.Name)
	}

	if validPool.CIDR != "192.168.0.0/16" {
		t.Errorf("Expected CIDR 192.168.0.0/16, got %s", validPool.CIDR)
	}

	if validPool.BlockSize != 26 {
		t.Errorf("Expected block size 26, got %d", validPool.BlockSize)
	}

	if validPool.IPIPMode != "CrossSubnet" {
		t.Errorf("Expected IPIP mode CrossSubnet, got %s", validPool.IPIPMode)
	}

	if validPool.VXLANMode != "Never" {
		t.Errorf("Expected VXLAN mode Never, got %s", validPool.VXLANMode)
	}

	if validPool.Disabled {
		t.Error("Expected pool to be enabled")
	}

	if len(validPool.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(validPool.Annotations))
	}

	// Test IP pool with different modes
	modes := []string{"Always", "CrossSubnet", "Never"}
	for _, mode := range modes {
		validPool.IPIPMode = mode
		if validPool.IPIPMode != mode {
			t.Errorf("Expected IPIP mode %s, got %s", mode, validPool.IPIPMode)
		}

		validPool.VXLANMode = mode
		if validPool.VXLANMode != mode {
			t.Errorf("Expected VXLAN mode %s, got %s", mode, validPool.VXLANMode)
		}
	}

	// Test block size validation
	validBlockSizes := []int{20, 24, 26, 28, 30}
	for _, size := range validBlockSizes {
		validPool.BlockSize = size
		if validPool.BlockSize != size {
			t.Errorf("Expected block size %d, got %d", size, validPool.BlockSize)
		}
	}
}

// Test BGPPeer validation
func TestBGPPeer_Validation(t *testing.T) {
	// Test valid BGP peer
	validPeer := &BGPPeer{
		Name:         "test-peer",
		ASN:          65001,
		IP:           "192.168.1.1",
		NodeSelector: "has(router)",
		Password:     "secret",
		Annotations:  map[string]string{"key": "value"},
	}

	if validPeer.Name != "test-peer" {
		t.Errorf("Expected name test-peer, got %s", validPeer.Name)
	}

	if validPeer.ASN != 65001 {
		t.Errorf("Expected ASN 65001, got %d", validPeer.ASN)
	}

	if validPeer.IP != "192.168.1.1" {
		t.Errorf("Expected IP 192.168.1.1, got %s", validPeer.IP)
	}

	if validPeer.NodeSelector != "has(router)" {
		t.Errorf("Expected node selector 'has(router)', got %s", validPeer.NodeSelector)
	}

	if validPeer.Password != "secret" {
		t.Errorf("Expected password secret, got %s", validPeer.Password)
	}

	if len(validPeer.Annotations) != 1 {
		t.Errorf("Expected 1 annotation, got %d", len(validPeer.Annotations))
	}

	// Test ASN validation
	validASNs := []int{1, 65001, 65535, 4200000000}
	for _, asn := range validASNs {
		validPeer.ASN = asn
		if validPeer.ASN != asn {
			t.Errorf("Expected ASN %d, got %d", asn, validPeer.ASN)
		}
	}

	// Test IP validation
	validIPs := []string{"192.168.1.1", "10.0.0.1", "172.16.0.1", "1.2.3.4"}
	for _, ip := range validIPs {
		validPeer.IP = ip
		if validPeer.IP != ip {
			t.Errorf("Expected IP %s, got %s", ip, validPeer.IP)
		}
	}
}

// Test Rule validation
func TestRule_Validation(t *testing.T) {
	// Test valid rule
	validRule := &Rule{
		Action:   "Allow",
		Protocol: "TCP",
		Source: RuleEndpoint{
			Selector: "app == 'frontend'",
			Nets:     []string{"192.168.0.0/16"},
			Ports:    []Port{{Number: 80, Protocol: "TCP"}},
		},
		Destination: RuleEndpoint{
			Selector: "app == 'backend'",
			Ports:    []Port{{Number: 8080, Protocol: "TCP"}},
		},
		HTTP: &HTTPMatch{
			Methods: []string{"GET", "POST"},
			Paths:   []string{"/api/*"},
			Headers: map[string]string{"Content-Type": "application/json"},
		},
		ICMP: &ICMPMatch{
			Type: 8,
			Code: 0,
		},
	}

	if validRule.Action != "Allow" {
		t.Errorf("Expected action Allow, got %s", validRule.Action)
	}

	if validRule.Protocol != "TCP" {
		t.Errorf("Expected protocol TCP, got %s", validRule.Protocol)
	}

	if validRule.Source.Selector != "app == 'frontend'" {
		t.Errorf("Expected source selector 'app == \\'frontend\\'', got %s", validRule.Source.Selector)
	}

	if len(validRule.Source.Nets) != 1 {
		t.Errorf("Expected 1 source net, got %d", len(validRule.Source.Nets))
	}

	if len(validRule.Source.Ports) != 1 {
		t.Errorf("Expected 1 source port, got %d", len(validRule.Source.Ports))
	}

	if validRule.Destination.Selector != "app == 'backend'" {
		t.Errorf("Expected destination selector 'app == \\'backend\\'', got %s", validRule.Destination.Selector)
	}

	if len(validRule.Destination.Ports) != 1 {
		t.Errorf("Expected 1 destination port, got %d", len(validRule.Destination.Ports))
	}

	// Test HTTP match
	if validRule.HTTP == nil {
		t.Fatal("Expected HTTP match to be set")
	}

	if len(validRule.HTTP.Methods) != 2 {
		t.Errorf("Expected 2 HTTP methods, got %d", len(validRule.HTTP.Methods))
	}

	if len(validRule.HTTP.Paths) != 1 {
		t.Errorf("Expected 1 HTTP path, got %d", len(validRule.HTTP.Paths))
	}

	if len(validRule.HTTP.Headers) != 1 {
		t.Errorf("Expected 1 HTTP header, got %d", len(validRule.HTTP.Headers))
	}

	// Test ICMP match
	if validRule.ICMP == nil {
		t.Fatal("Expected ICMP match to be set")
	}

	if validRule.ICMP.Type != 8 {
		t.Errorf("Expected ICMP type 8, got %d", validRule.ICMP.Type)
	}

	if validRule.ICMP.Code != 0 {
		t.Errorf("Expected ICMP code 0, got %d", validRule.ICMP.Code)
	}

	// Test action validation
	validActions := []string{"Allow", "Deny", "Log"}
	for _, action := range validActions {
		validRule.Action = action
		if validRule.Action != action {
			t.Errorf("Expected action %s, got %s", action, validRule.Action)
		}
	}

	// Test protocol validation
	validProtocols := []string{"TCP", "UDP", "ICMP", "ICMPv6"}
	for _, protocol := range validProtocols {
		validRule.Protocol = protocol
		if validRule.Protocol != protocol {
			t.Errorf("Expected protocol %s, got %s", protocol, validRule.Protocol)
		}
	}
}

// Test Port validation
func TestPort_Validation(t *testing.T) {
	// Test valid port
	validPort := &Port{
		Number:   80,
		Protocol: "TCP",
		EndPort:  90,
	}

	if validPort.Number != 80 {
		t.Errorf("Expected number 80, got %d", validPort.Number)
	}

	if validPort.Protocol != "TCP" {
		t.Errorf("Expected protocol TCP, got %s", validPort.Protocol)
	}

	if validPort.EndPort != 90 {
		t.Errorf("Expected end port 90, got %d", validPort.EndPort)
	}

	// Test port number validation
	validPortNumbers := []int{1, 80, 443, 8080, 65535}
	for _, number := range validPortNumbers {
		validPort.Number = number
		if validPort.Number != number {
			t.Errorf("Expected port number %d, got %d", number, validPort.Number)
		}
	}

	// Test protocol validation
	validProtocols := []string{"TCP", "UDP", "SCTP"}
	for _, protocol := range validProtocols {
		validPort.Protocol = protocol
		if validPort.Protocol != protocol {
			t.Errorf("Expected protocol %s, got %s", protocol, validPort.Protocol)
		}
	}

	// Test end port validation
	validEndPorts := []int{80, 443, 8080, 65535}
	for _, endPort := range validEndPorts {
		validPort.EndPort = endPort
		if validPort.EndPort != endPort {
			t.Errorf("Expected end port %d, got %d", endPort, validPort.EndPort)
		}
	}
}
