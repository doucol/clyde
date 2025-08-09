// Package calico provides comprehensive Calico OSS management capabilities
package calico

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GetFelixConfiguration returns the Felix configuration
func (cm *CalicoManager) GetFelixConfiguration(ctx context.Context) (*FelixConfiguration, error) {
	cm.Logf("Getting Felix configuration...")

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "felixconfigurations",
	}

	list, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list Felix configurations: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no Felix configuration found")
	}

	item := list.Items[0]
	spec := item.Object["spec"].(map[string]any)

	config := &FelixConfiguration{
		Name: item.GetName(),
	}

	if logSeverity, ok := spec["logSeverityScreen"].(string); ok {
		config.LogSeverityScreen = logSeverity
	}
	if logSeverity, ok := spec["logSeverityFile"].(string); ok {
		config.LogSeverityFile = logSeverity
	}
	if logSeverity, ok := spec["logSeveritySys"].(string); ok {
		config.LogSeveritySys = logSeverity
	}
	if enabled, ok := spec["prometheusMetricsEnabled"].(bool); ok {
		config.PrometheusMetricsEnabled = enabled
	}
	if port, ok := spec["prometheusMetricsPort"].(float64); ok {
		config.PrometheusMetricsPort = int(port)
	}
	if enabled, ok := spec["prometheusGoMetricsEnabled"].(bool); ok {
		config.PrometheusGoMetricsEnabled = enabled
	}
	if enabled, ok := spec["prometheusProcessMetricsEnabled"].(bool); ok {
		config.PrometheusProcessMetricsEnabled = enabled
	}

	return config, nil
}

// UpdateFelixConfiguration updates the Felix configuration
// func (cm *CalicoManager) UpdateFelixConfiguration(ctx context.Context, config *FelixConfiguration) error {
// 	cm.Logf("Updating Felix configuration...")
//
// 	manifest := cm.generateFelixConfigurationManifest(config)
//
// 	if err := cm.applyManifest(ctx, manifest); err != nil {
// 		return fmt.Errorf("failed to update Felix configuration: %w", err)
// 	}
//
// 	cm.Logf("Felix configuration updated successfully")
// 	return nil
// }

// GetBGPConfiguration returns the BGP configuration
func (cm *CalicoManager) GetBGPConfiguration(ctx context.Context) (*BGPConfiguration, error) {
	cm.Logf("Getting BGP configuration...")

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "bgpconfigurations",
	}

	list, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list BGP configurations: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no BGP configuration found")
	}

	item := list.Items[0]
	spec := item.Object["spec"].(map[string]interface{})

	config := &BGPConfiguration{
		Name: item.GetName(),
	}

	if asNumber, ok := spec["asNumber"].(float64); ok {
		config.ASNumber = int(asNumber)
	}
	if serviceClusterIPs, ok := spec["serviceClusterIPs"].([]interface{}); ok {
		for _, ip := range serviceClusterIPs {
			config.ServiceClusterIPs = append(config.ServiceClusterIPs, ip.(string))
		}
	}
	if serviceExternalIPs, ok := spec["serviceExternalIPs"].([]interface{}); ok {
		for _, ip := range serviceExternalIPs {
			config.ServiceExternalIPs = append(config.ServiceExternalIPs, ip.(string))
		}
	}
	if serviceLoadBalancerIPs, ok := spec["serviceLoadBalancerIPs"].([]interface{}); ok {
		for _, ip := range serviceLoadBalancerIPs {
			config.ServiceLoadBalancerIPs = append(config.ServiceLoadBalancerIPs, ip.(string))
		}
	}

	return config, nil
}

// UpdateBGPConfiguration updates the BGP configuration
// func (cm *CalicoManager) UpdateBGPConfiguration(ctx context.Context, config *BGPConfiguration) error {
// 	cm.Logf("Updating BGP configuration...")
//
// 	manifest := cm.generateBGPConfigurationManifest(config)
//
// 	if err := cm.applyManifest(ctx, manifest); err != nil {
// 		return fmt.Errorf("failed to update BGP configuration: %w", err)
// 	}
//
// 	cm.Logf("BGP configuration updated successfully")
// 	return nil
// }

// GetClusterInformation returns cluster information
func (cm *CalicoManager) GetClusterInformation(ctx context.Context) (*ClusterInformation, error) {
	cm.Logf("Getting cluster information...")

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "clusterinformations",
	}

	list, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list cluster information: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no cluster information found")
	}

	item := list.Items[0]
	spec := item.Object["spec"].(map[string]interface{})

	info := &ClusterInformation{
		Name: item.GetName(),
	}

	if calicoVersion, ok := spec["calicoVersion"].(string); ok {
		info.CalicoVersion = calicoVersion
	}
	if clusterType, ok := spec["clusterType"].(string); ok {
		info.ClusterType = clusterType
	}
	if datastoreType, ok := spec["datastoreType"].(string); ok {
		info.DatastoreType = datastoreType
	}

	return info, nil
}

// GetGlobalNetworkSets returns all global network sets
func (cm *CalicoManager) GetGlobalNetworkSets(ctx context.Context) ([]*GlobalNetworkSet, error) {
	cm.Logf("Getting global network sets...")

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "globalnetworksets",
	}

	list, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list global network sets: %w", err)
	}

	var networkSets []*GlobalNetworkSet
	for _, item := range list.Items {
		spec := item.Object["spec"].(map[string]interface{})

		networkSet := &GlobalNetworkSet{
			Name: item.GetName(),
		}

		if nets, ok := spec["nets"].([]interface{}); ok {
			for _, net := range nets {
				networkSet.Nets = append(networkSet.Nets, net.(string))
			}
		}

		networkSets = append(networkSets, networkSet)
	}

	return networkSets, nil
}

// CreateGlobalNetworkSet creates a new global network set
// func (cm *CalicoManager) CreateGlobalNetworkSet(ctx context.Context, networkSet *GlobalNetworkSet) error {
// 	cm.Logf("Creating global network set %s", networkSet.Name)
//
// 	manifest := cm.generateGlobalNetworkSetManifest(networkSet)
//
// 	if err := cm.applyManifest(ctx, manifest); err != nil {
// 		return fmt.Errorf("failed to create global network set: %w", err)
// 	}
//
// 	cm.Logf("Global network set %s created successfully", networkSet.Name)
// 	return nil
// }

// DeleteGlobalNetworkSet deletes a global network set
func (cm *CalicoManager) DeleteGlobalNetworkSet(ctx context.Context, name string) error {
	cm.Logf("Deleting global network set %s", name)

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "globalnetworksets",
	}

	if err := cm.dynamic.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to delete global network set: %w", err)
	}

	cm.Logf("Global network set %s deleted successfully", name)
	return nil
}

// GetHostEndpoints returns all host endpoints
func (cm *CalicoManager) GetHostEndpoints(ctx context.Context) ([]*HostEndpoint, error) {
	cm.Logf("Getting host endpoints...")

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "hostendpoints",
	}

	list, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list host endpoints: %w", err)
	}

	var endpoints []*HostEndpoint
	for _, item := range list.Items {
		spec := item.Object["spec"].(map[string]interface{})

		endpoint := &HostEndpoint{
			Name: item.GetName(),
		}

		if node, ok := spec["node"].(string); ok {
			endpoint.Node = node
		}
		if iface, ok := spec["interfaceName"].(string); ok {
			endpoint.Interface = iface
		}
		if expectedIPs, ok := spec["expectedIPs"].([]interface{}); ok {
			for _, ip := range expectedIPs {
				endpoint.ExpectedIPs = append(endpoint.ExpectedIPs, ip.(string))
			}
		}
		if profiles, ok := spec["profiles"].([]interface{}); ok {
			for _, profile := range profiles {
				endpoint.Profiles = append(endpoint.Profiles, profile.(string))
			}
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}

// CreateHostEndpoint creates a new host endpoint
// func (cm *CalicoManager) CreateHostEndpoint(ctx context.Context, endpoint *HostEndpoint) error {
// 	cm.Logf("Creating host endpoint %s", endpoint.Name)
//
// 	manifest := cm.generateHostEndpointManifest(endpoint)
//
// 	if err := cm.applyManifest(ctx, manifest); err != nil {
// 		return fmt.Errorf("failed to create host endpoint: %w", err)
// 	}
//
// 	cm.Logf("Host endpoint %s created successfully", endpoint.Name)
// 	return nil
// }

// DeleteHostEndpoint deletes a host endpoint
func (cm *CalicoManager) DeleteHostEndpoint(ctx context.Context, name string) error {
	cm.Logf("Deleting host endpoint %s", name)

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "hostendpoints",
	}

	if err := cm.dynamic.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed to delete host endpoint: %w", err)
	}

	cm.Logf("Host endpoint %s deleted successfully", name)
	return nil
}

// GetWorkloadEndpoints returns workload endpoints for a namespace
func (cm *CalicoManager) GetWorkloadEndpoints(ctx context.Context, namespace string) ([]*WorkloadEndpoint, error) {
	cm.Logf("Getting workload endpoints for namespace %s", namespace)

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "workloadendpoints",
	}

	list, err := cm.dynamic.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list workload endpoints: %w", err)
	}

	var endpoints []*WorkloadEndpoint
	for _, item := range list.Items {
		spec := item.Object["spec"].(map[string]interface{})

		endpoint := &WorkloadEndpoint{
			Name:      item.GetName(),
			Namespace: item.GetNamespace(),
		}

		if pod, ok := spec["pod"].(string); ok {
			endpoint.Pod = pod
		}
		if container, ok := spec["containerID"].(string); ok {
			endpoint.Container = container
		}
		if iface, ok := spec["interfaceName"].(string); ok {
			endpoint.Interface = iface
		}
		if expectedIPs, ok := spec["ipNetworks"].([]interface{}); ok {
			for _, ip := range expectedIPs {
				endpoint.ExpectedIPs = append(endpoint.ExpectedIPs, ip.(string))
			}
		}
		if profiles, ok := spec["profiles"].([]interface{}); ok {
			for _, profile := range profiles {
				endpoint.Profiles = append(endpoint.Profiles, profile.(string))
			}
		}

		endpoints = append(endpoints, endpoint)
	}

	return endpoints, nil
}
