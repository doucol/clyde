// Package calico provides comprehensive Calico OSS management capabilities
package calico

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Example demonstrates how to use the Calico package
func Example() {
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
	calicoManager := NewCalicoManager(clientset, dynamicClient, config, os.Stdout)

	ctx := context.Background()

	// Example 1: Install Calico
	installOptions := &InstallOptions{
		Version:          "v3.26.1",
		InstallationType: "operator",
		CNI:              "calico",
		IPAM:             "calico-ipam",
		Datastore:        "kubernetes",
		EnablePrometheus: true,
	}

	err = calicoManager.Install(ctx, installOptions)
	if err != nil {
		log.Printf("Failed to install Calico: %v", err)
	}

	// Example 2: Wait for Calico to be ready
	err = calicoManager.WaitForReady(ctx, 5*time.Minute)
	if err != nil {
		log.Printf("Failed to wait for Calico ready: %v", err)
	}

	// Example 3: Get Calico status
	status, err := calicoManager.GetStatus(ctx)
	if err != nil {
		log.Printf("Failed to get status: %v", err)
	} else {
		fmt.Printf("Calico installed: %v, Version: %s\n", status.Installed, status.Version)
	}

	// Example 4: Create an IP pool
	ipPool := &IPPool{
		Name:      "example-pool",
		CIDR:      "192.168.0.0/16",
		BlockSize: 26,
		IPIPMode:  "CrossSubnet",
		VXLANMode: "Never",
	}

	err = calicoManager.CreateIPPool(ctx, ipPool)
	if err != nil {
		log.Printf("Failed to create IP pool: %v", err)
	}

	// Example 5: Create a network policy
	networkPolicy := &NetworkPolicy{
		Name:      "example-policy",
		Namespace: "default",
		Selector:  "app == 'web'",
		Types:     []string{"Ingress", "Egress"},
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
	}

	err = calicoManager.CreateNetworkPolicy(ctx, networkPolicy)
	if err != nil {
		log.Printf("Failed to create network policy: %v", err)
	}

	// Example 6: Create a BGP peer
	bgpPeer := &BGPPeer{
		Name:         "example-peer",
		ASN:          65001,
		IP:           "192.168.1.1",
		NodeSelector: "has(router)",
	}

	err = calicoManager.CreateBGPPeer(ctx, bgpPeer)
	if err != nil {
		log.Printf("Failed to create BGP peer: %v", err)
	}

	// Example 7: Get Felix configuration
	felixConfig, err := calicoManager.GetFelixConfiguration(ctx)
	if err != nil {
		log.Printf("Failed to get Felix configuration: %v", err)
	} else {
		fmt.Printf("Felix log severity: %s\n", felixConfig.LogSeverityScreen)
	}

	// Example 8: Update Felix configuration
	if felixConfig != nil {
		felixConfig.LogSeverityScreen = "Info"
		felixConfig.PrometheusMetricsEnabled = true
		felixConfig.PrometheusMetricsPort = 9091

		err = calicoManager.UpdateFelixConfiguration(ctx, felixConfig)
		if err != nil {
			log.Printf("Failed to update Felix configuration: %v", err)
		}
	}

	// Example 9: Get BGP configuration
	bgpConfig, err := calicoManager.GetBGPConfiguration(ctx)
	if err != nil {
		log.Printf("Failed to get BGP configuration: %v", err)
	} else {
		fmt.Printf("BGP ASN: %d\n", bgpConfig.ASNumber)
	}

	// Example 10: Create a global network set
	networkSet := &GlobalNetworkSet{
		Name: "example-networks",
		Nets: []string{
			"10.0.0.0/8",
			"172.16.0.0/12",
		},
		Labels: map[string]string{
			"environment": "production",
		},
	}

	err = calicoManager.CreateGlobalNetworkSet(ctx, networkSet)
	if err != nil {
		log.Printf("Failed to create global network set: %v", err)
	}

	// Example 11: Create a host endpoint
	hostEndpoint := &HostEndpoint{
		Name:      "example-host",
		Node:      "node-1",
		Interface: "eth0",
		ExpectedIPs: []string{
			"192.168.1.10",
		},
		Profiles: []string{
			"default",
		},
	}

	err = calicoManager.CreateHostEndpoint(ctx, hostEndpoint)
	if err != nil {
		log.Printf("Failed to create host endpoint: %v", err)
	}

	// Example 12: Health check
	healthResult, err := calicoManager.HealthCheck(ctx)
	if err != nil {
		log.Printf("Failed to perform health check: %v", err)
	} else {
		fmt.Printf("Calico health: %v, Errors: %d\n", healthResult.Overall, len(healthResult.Errors))
	}

	// Example 13: Apply custom YAML resource
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

	err = calicoManager.ApplyResource(ctx, customYAML)
	if err != nil {
		log.Printf("Failed to apply custom resource: %v", err)
	}

	// Example 14: Upgrade Calico
	upgradeOptions := &UpgradeOptions{
		Version:              "v3.27.0",
		BackupBeforeUpgrade:  true,
		ValidateAfterUpgrade: true,
	}

	err = calicoManager.Upgrade(ctx, upgradeOptions)
	if err != nil {
		log.Printf("Failed to upgrade Calico: %v", err)
	}

	// Example 15: Uninstall Calico (commented out for safety)
	/*
		uninstallOptions := &UninstallOptions{
			RemoveCRDs:      true,
			RemoveNamespace:  true,
			RemoveFinalizers: true,
		}

		err = calicoManager.Uninstall(ctx, uninstallOptions)
		if err != nil {
			log.Printf("Failed to uninstall Calico: %v", err)
		}
	*/
}
