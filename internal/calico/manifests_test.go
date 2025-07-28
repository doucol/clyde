package calico

import (
	"strings"
	"testing"
)

// Test generateOperatorManifests
func TestCalicoManager_GenerateOperatorManifests(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	options := &InstallOptions{
		Version:          "v3.26.1",
		InstallationType: "operator",
		CNI:              "calico",
		IPAM:             "calico-ipam",
		Datastore:        "kubernetes",
		EnablePrometheus: true,
	}

	manifests := manager.generateOperatorManifests(options)

	if len(manifests) == 0 {
		t.Fatal("Expected operator manifests to be generated")
	}

	// Check that CRD manifest is generated
	foundCRD := false
	for _, manifest := range manifests {
		if strings.Contains(manifest, "CustomResourceDefinition") {
			foundCRD = true
			break
		}
	}
	if !foundCRD {
		t.Error("Expected CRD manifest to be generated")
	}

	// Check that operator deployment manifest is generated
	foundDeployment := false
	for _, manifest := range manifests {
		if strings.Contains(manifest, "Deployment") && strings.Contains(manifest, "tigera-operator") {
			foundDeployment = true
			break
		}
	}
	if !foundDeployment {
		t.Error("Expected operator deployment manifest to be generated")
	}

	// Check that version is included in deployment
	foundVersion := false
	for _, manifest := range manifests {
		if strings.Contains(manifest, "tigera/operator:v3.26.1") {
			foundVersion = true
			break
		}
	}
	if !foundVersion {
		t.Error("Expected version to be included in deployment manifest")
	}
}

// Test generateCalicoInstanceManifest
func TestCalicoManager_GenerateCalicoInstanceManifest(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	options := &InstallOptions{
		Version:          "v3.26.1",
		InstallationType: "operator",
		CNI:              "calico",
		IPAM:             "calico-ipam",
		Datastore:        "kubernetes",
		EnablePrometheus: true,
	}

	manifest := manager.generateCalicoInstanceManifest(options)

	if manifest == "" {
		t.Fatal("Expected Calico instance manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: operator.tigera.io/v1",
		"kind: Installation",
		"name: default",
		"type: calico",
		"cidr: 10.42.0.0/16",
		"blockSize: 26",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}
}

// Test generateIPPoolManifest
func TestCalicoManager_GenerateIPPoolManifest(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	pool := &IPPool{
		Name:        "test-pool",
		CIDR:        "192.168.0.0/16",
		BlockSize:   26,
		IPIPMode:    "CrossSubnet",
		VXLANMode:   "Never",
		Disabled:    false,
		Annotations: map[string]string{"key": "value"},
	}

	manifest := manager.generateIPPoolManifest(pool)

	if manifest == "" {
		t.Fatal("Expected IP pool manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: crd.projectcalico.org/v1",
		"kind: IPPool",
		"name: test-pool",
		"cidr: 192.168.0.0/16",
		"blockSize: 26",
		"ipipMode: CrossSubnet",
		"vxlanMode: Never",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}
}

// Test generateNetworkPolicyManifest
func TestCalicoManager_GenerateNetworkPolicyManifest(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	policy := &NetworkPolicy{
		Name:         "test-policy",
		Namespace:    "default",
		Selector:     "app == 'web'",
		IngressRules: []Rule{},
		EgressRules:  []Rule{},
		Types:        []string{"Ingress", "Egress"},
		Annotations:  map[string]string{"key": "value"},
	}

	manifest := manager.generateNetworkPolicyManifest(policy)

	if manifest == "" {
		t.Fatal("Expected network policy manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: crd.projectcalico.org/v1",
		"kind: NetworkPolicy",
		"name: test-policy",
		"namespace: default",
		"selector: app == 'web'",
		"- Ingress",
		"- Egress",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}
}

// Test generateNetworkPolicyManifest with rules
func TestCalicoManager_GenerateNetworkPolicyManifest_WithRules(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	policy := &NetworkPolicy{
		Name:      "test-policy",
		Namespace: "default",
		Selector:  "app == 'web'",
		IngressRules: []Rule{
			{
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
			},
		},
		EgressRules: []Rule{
			{
				Action:   "Deny",
				Protocol: "UDP",
				Source: RuleEndpoint{
					Nets: []string{"192.168.1.0/24"},
				},
			},
		},
		Types:       []string{"Ingress", "Egress"},
		Annotations: map[string]string{"key": "value"},
	}

	manifest := manager.generateNetworkPolicyManifest(policy)

	if manifest == "" {
		t.Fatal("Expected network policy manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: crd.projectcalico.org/v1",
		"kind: NetworkPolicy",
		"name: test-policy",
		"namespace: default",
		"selector: app == 'web'",
		"- Ingress",
		"- Egress",
		"ingress:",
		"egress:",
		"action: Allow",
		"action: Deny",
		"protocol: TCP",
		"protocol: UDP",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}
}

// Test generateBGPPeerManifest
func TestCalicoManager_GenerateBGPPeerManifest(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	peer := &BGPPeer{
		Name:         "test-peer",
		ASN:          65001,
		IP:           "192.168.1.1",
		NodeSelector: "has(router)",
		Password:     "secret",
		Annotations:  map[string]string{"key": "value"},
	}

	manifest := manager.generateBGPPeerManifest(peer)

	if manifest == "" {
		t.Fatal("Expected BGP peer manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: crd.projectcalico.org/v1",
		"kind: BGPPeer",
		"name: test-peer",
		"asNumber: 65001",
		"peerIP: 192.168.1.1",
		"nodeSelector: has(router)",
		"password: secret",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}
}

// Test generateBGPPeerManifest without optional fields
func TestCalicoManager_GenerateBGPPeerManifest_Minimal(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	peer := &BGPPeer{
		Name: "test-peer",
		ASN:  65001,
		IP:   "192.168.1.1",
	}

	manifest := manager.generateBGPPeerManifest(peer)

	if manifest == "" {
		t.Fatal("Expected BGP peer manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: crd.projectcalico.org/v1",
		"kind: BGPPeer",
		"name: test-peer",
		"asNumber: 65001",
		"peerIP: 192.168.1.1",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}

	// Check that optional fields are not present
	optionalFields := []string{
		"nodeSelector:",
		"password:",
	}

	for _, field := range optionalFields {
		if strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to not contain: %s", field)
		}
	}
}

// Test generateFelixConfigurationManifest
func TestCalicoManager_GenerateFelixConfigurationManifest(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

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

	manifest := manager.generateFelixConfigurationManifest(config)

	if manifest == "" {
		t.Fatal("Expected Felix configuration manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: crd.projectcalico.org/v1",
		"kind: FelixConfiguration",
		"name: default",
		"logSeverityScreen: Info",
		"logSeverityFile: Info",
		"logSeveritySys: Info",
		"prometheusMetricsEnabled: true",
		"prometheusMetricsPort: 9091",
		"prometheusGoMetricsEnabled: true",
		"prometheusProcessMetricsEnabled: true",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}
}

// Test generateBGPConfigurationManifest
func TestCalicoManager_GenerateBGPConfigurationManifest(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	config := &BGPConfiguration{
		Name:                   "default",
		ASNumber:               65001,
		ServiceClusterIPs:      []string{"10.96.0.0/12"},
		ServiceExternalIPs:     []string{"1.2.3.4/32"},
		ServiceLoadBalancerIPs: []string{"5.6.7.8/32"},
	}

	manifest := manager.generateBGPConfigurationManifest(config)

	if manifest == "" {
		t.Fatal("Expected BGP configuration manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: crd.projectcalico.org/v1",
		"kind: BGPConfiguration",
		"name: default",
		"asNumber: 65001",
		"serviceClusterIPs:",
		"- 10.96.0.0/12",
		"serviceExternalIPs:",
		"- 1.2.3.4/32",
		"serviceLoadBalancerIPs:",
		"- 5.6.7.8/32",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}
}

// Test generateGlobalNetworkSetManifest
func TestCalicoManager_GenerateGlobalNetworkSetManifest(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	networkSet := &GlobalNetworkSet{
		Name:        "test-networks",
		Nets:        []string{"10.0.0.0/8", "172.16.0.0/12"},
		Labels:      map[string]string{"environment": "production"},
		Annotations: map[string]string{"key": "value"},
	}

	manifest := manager.generateGlobalNetworkSetManifest(networkSet)

	if manifest == "" {
		t.Fatal("Expected global network set manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: crd.projectcalico.org/v1",
		"kind: GlobalNetworkSet",
		"name: test-networks",
		"nets:",
		"- 10.0.0.0/8",
		"- 172.16.0.0/12",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}
}

// Test generateHostEndpointManifest
func TestCalicoManager_GenerateHostEndpointManifest(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	endpoint := &HostEndpoint{
		Name:        "test-host",
		Node:        "node-1",
		Interface:   "eth0",
		ExpectedIPs: []string{"192.168.1.10"},
		Profiles:    []string{"default"},
		Labels:      map[string]string{"environment": "production"},
		Annotations: map[string]string{"key": "value"},
	}

	manifest := manager.generateHostEndpointManifest(endpoint)

	if manifest == "" {
		t.Fatal("Expected host endpoint manifest to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"apiVersion: crd.projectcalico.org/v1",
		"kind: HostEndpoint",
		"name: test-host",
		"node: node-1",
		"interfaceName: eth0",
		"expectedIPs:",
		"- 192.168.1.10",
		"profiles:",
		"- default",
	}

	for _, field := range requiredFields {
		if !strings.Contains(manifest, field) {
			t.Errorf("Expected manifest to contain: %s", field)
		}
	}
}

// Test generateRuleYAML
func TestCalicoManager_GenerateRuleYAML(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	rule := Rule{
		Action:   "Allow",
		Protocol: "TCP",
		Source: RuleEndpoint{
			Selector: "app == 'frontend'",
			Nets:     []string{"192.168.0.0/16"},
		},
		Destination: RuleEndpoint{
			Ports: []Port{
				{Number: 80, Protocol: "TCP"},
				{Number: 443, Protocol: "TCP"},
			},
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

	yaml := manager.generateRuleYAML(rule, "  ")

	if yaml == "" {
		t.Fatal("Expected rule YAML to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"action: Allow",
		"protocol: TCP",
		"source:",
		"selector: app == 'frontend'",
		"nets:",
		"- 192.168.0.0/16",
		"destination:",
		"ports:",
		"- port: 80",
		"protocol: TCP",
		"- port: 443",
		"http:",
		"methods:",
		"- GET",
		"- POST",
		"paths:",
		"- /api/*",
		"icmp:",
		"type: 8",
		"code: 0",
	}

	for _, field := range requiredFields {
		if !strings.Contains(yaml, field) {
			t.Errorf("Expected YAML to contain: %s", field)
		}
	}
}

// Test generateRuleYAML with minimal rule
func TestCalicoManager_GenerateRuleYAML_Minimal(t *testing.T) {
	manager := NewCalicoManager(nil, nil, nil, nil)

	rule := Rule{
		Action: "Allow",
	}

	yaml := manager.generateRuleYAML(rule, "  ")

	if yaml == "" {
		t.Fatal("Expected rule YAML to be generated")
	}

	// Check for required fields
	requiredFields := []string{
		"action: Allow",
	}

	for _, field := range requiredFields {
		if !strings.Contains(yaml, field) {
			t.Errorf("Expected YAML to contain: %s", field)
		}
	}

	// Check that optional fields are not present
	optionalFields := []string{
		"protocol:",
		"source:",
		"destination:",
		"http:",
		"icmp:",
	}

	for _, field := range optionalFields {
		if strings.Contains(yaml, field) {
			t.Errorf("Expected YAML to not contain: %s", field)
		}
	}
} 