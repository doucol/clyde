// Package calico provides comprehensive Calico OSS management capabilities
package calico

// // generateOperatorManifests generates operator manifests based on options
// func (cm *CalicoManager) generateOperatorManifests(options *InstallOptions) []string {
// 	var manifests []string
//
// 	// Generate CRDs manifest
// 	crdManifest := `apiVersion: apiextensions.k8s.io/v1
// kind: CustomResourceDefinition
// metadata:
//   name: installations.operator.tigera.io
// spec:
//   group: operator.tigera.io
//   names:
//     kind: Installation
//     listKind: InstallationList
//     plural: installations
//     singular: installation
//   scope: Cluster
//   versions:
//   - name: v1
//     schema:
//       openAPIV3Schema:
//         type: object
//     served: true
//     storage: true
//     subresources:
//       status: {}
// `
// 	manifests = append(manifests, crdManifest)
//
// 	// Generate operator deployment manifest
// 	operatorManifest := fmt.Sprintf(`apiVersion: apps/v1
// kind: Deployment
// metadata:
//   name: tigera-operator
//   namespace: tigera-operator
// spec:
//   replicas: 1
//   selector:
//     matchLabels:
//       name: tigera-operator
//   template:
//     metadata:
//       labels:
//         name: tigera-operator
//     spec:
//       containers:
//       - name: tigera-operator
//         image: tigera/operator:v%s
//         env:
//         - name: WATCH_NAMESPACE
//           value: ""
//         - name: POD_NAME
//           valueFrom:
//             fieldRef:
//               fieldPath: metadata.name
//         - name: OPERATOR_NAME
//           value: tigera-operator
// `, options.Version)
//
// 	manifests = append(manifests, operatorManifest)
//
// 	return manifests
// }
//
// // generateCalicoInstanceManifest generates Calico instance manifest
// func (cm *CalicoManager) generateCalicoInstanceManifest(options *InstallOptions) string {
// 	manifest := fmt.Sprintf(`apiVersion: operator.tigera.io/v1
// kind: Installation
// metadata:
//   name: default
// spec:
//   cni:
//     type: %s
//   ipPools:
//   - blockSize: 26
//     cidr: 10.42.0.0/16
//     encapsulation: VXLANCrossSubnet
//     natOutgoing: true
//     nodeSelector: all()
//   typhaMetricsPort: 9093
//   nodeMetricsPort: 9091
//   flexVolumePath: /usr/libexec/kubernetes/kubelet-plugins/volume/exec/
//   nodeUpdateStrategy:
//     type: RollingUpdate
//     rollingUpdate:
//       maxUnavailable: 1
//   componentResources:
//   - componentName: Node
//     resourceRequirements:
//       limits:
//         cpu: 250m
//         memory: 500Mi
//       requests:
//         cpu: 250m
//         memory: 500Mi
//   - componentName: Typha
//     resourceRequirements:
//       limits:
//         cpu: 1000m
//         memory: 500Mi
//       requests:
//         cpu: 1000m
//         memory: 500Mi
//   - componentName: KubeControllers
//     resourceRequirements:
//       limits:
//         cpu: 1000m
//         memory: 500Mi
//       requests:
//         cpu: 1000m
//         memory: 500Mi
// `, options.CNI)
//
// 	return manifest
// }
//
// // generateIPPoolManifest generates IP pool manifest
// func (cm *CalicoManager) generateIPPoolManifest(pool *IPPool) string {
// 	manifest := fmt.Sprintf(`apiVersion: crd.projectcalico.org/v1
// kind: IPPool
// metadata:
//   name: %s
// spec:
//   cidr: %s
//   blockSize: %d
//   ipipMode: %s
//   vxlanMode: %s
//   natOutgoing: true
//   nodeSelector: all()
// `, pool.Name, pool.CIDR, pool.BlockSize, pool.IPIPMode, pool.VXLANMode)
//
// 	return manifest
// }
//
// // generateNetworkPolicyManifest generates network policy manifest
// func (cm *CalicoManager) generateNetworkPolicyManifest(policy *NetworkPolicy) string {
// 	manifest := fmt.Sprintf(`apiVersion: crd.projectcalico.org/v1
// kind: NetworkPolicy
// metadata:
//   name: %s
//   namespace: %s
// spec:
//   selector: %s
//   types:
// `, policy.Name, policy.Namespace, policy.Selector)
//
// 	for _, policyType := range policy.Types {
// 		manifest += fmt.Sprintf("  - %s\n", policyType)
// 	}
//
// 	if len(policy.IngressRules) > 0 {
// 		manifest += "  ingress:\n"
// 		for _, rule := range policy.IngressRules {
// 			manifest += cm.generateRuleYAML(rule, "    ")
// 		}
// 	}
//
// 	if len(policy.EgressRules) > 0 {
// 		manifest += "  egress:\n"
// 		for _, rule := range policy.EgressRules {
// 			manifest += cm.generateRuleYAML(rule, "    ")
// 		}
// 	}
//
// 	return manifest
// }
//
// // generateBGPPeerManifest generates BGP peer manifest
// func (cm *CalicoManager) generateBGPPeerManifest(peer *BGPPeer) string {
// 	manifest := fmt.Sprintf(`apiVersion: crd.projectcalico.org/v1
// kind: BGPPeer
// metadata:
//   name: %s
// spec:
//   asNumber: %d
//   peerIP: %s
// `, peer.Name, peer.ASN, peer.IP)
//
// 	if peer.NodeSelector != "" {
// 		manifest += fmt.Sprintf("  nodeSelector: %s\n", peer.NodeSelector)
// 	}
//
// 	if peer.Password != "" {
// 		manifest += fmt.Sprintf("  password: %s\n", peer.Password)
// 	}
//
// 	return manifest
// }
//
// // generateFelixConfigurationManifest generates Felix configuration manifest
// func (cm *CalicoManager) generateFelixConfigurationManifest(config *FelixConfiguration) string {
// 	manifest := fmt.Sprintf(`apiVersion: crd.projectcalico.org/v1
// kind: FelixConfiguration
// metadata:
//   name: %s
// spec:
// `, config.Name)
//
// 	if config.LogSeverityScreen != "" {
// 		manifest += fmt.Sprintf("  logSeverityScreen: %s\n", config.LogSeverityScreen)
// 	}
// 	if config.LogSeverityFile != "" {
// 		manifest += fmt.Sprintf("  logSeverityFile: %s\n", config.LogSeverityFile)
// 	}
// 	if config.LogSeveritySys != "" {
// 		manifest += fmt.Sprintf("  logSeveritySys: %s\n", config.LogSeveritySys)
// 	}
// 	if config.PrometheusMetricsEnabled {
// 		manifest += "  prometheusMetricsEnabled: true\n"
// 	}
// 	if config.PrometheusMetricsPort > 0 {
// 		manifest += fmt.Sprintf("  prometheusMetricsPort: %d\n", config.PrometheusMetricsPort)
// 	}
// 	if config.PrometheusGoMetricsEnabled {
// 		manifest += "  prometheusGoMetricsEnabled: true\n"
// 	}
// 	if config.PrometheusProcessMetricsEnabled {
// 		manifest += "  prometheusProcessMetricsEnabled: true\n"
// 	}
//
// 	return manifest
// }
//
// // generateBGPConfigurationManifest generates BGP configuration manifest
// func (cm *CalicoManager) generateBGPConfigurationManifest(config *BGPConfiguration) string {
// 	manifest := fmt.Sprintf(`apiVersion: crd.projectcalico.org/v1
// kind: BGPConfiguration
// metadata:
//   name: %s
// spec:
//   asNumber: %d
// `, config.Name, config.ASNumber)
//
// 	if len(config.ServiceClusterIPs) > 0 {
// 		manifest += "  serviceClusterIPs:\n"
// 		for _, ip := range config.ServiceClusterIPs {
// 			manifest += fmt.Sprintf("  - %s\n", ip)
// 		}
// 	}
//
// 	if len(config.ServiceExternalIPs) > 0 {
// 		manifest += "  serviceExternalIPs:\n"
// 		for _, ip := range config.ServiceExternalIPs {
// 			manifest += fmt.Sprintf("  - %s\n", ip)
// 		}
// 	}
//
// 	if len(config.ServiceLoadBalancerIPs) > 0 {
// 		manifest += "  serviceLoadBalancerIPs:\n"
// 		for _, ip := range config.ServiceLoadBalancerIPs {
// 			manifest += fmt.Sprintf("  - %s\n", ip)
// 		}
// 	}
//
// 	return manifest
// }
//
// // generateGlobalNetworkSetManifest generates global network set manifest
// func (cm *CalicoManager) generateGlobalNetworkSetManifest(networkSet *GlobalNetworkSet) string {
// 	manifest := fmt.Sprintf(`apiVersion: crd.projectcalico.org/v1
// kind: GlobalNetworkSet
// metadata:
//   name: %s
// spec:
//   nets:
// `, networkSet.Name)
//
// 	for _, net := range networkSet.Nets {
// 		manifest += fmt.Sprintf("  - %s\n", net)
// 	}
//
// 	return manifest
// }
//
// // generateHostEndpointManifest generates host endpoint manifest
// func (cm *CalicoManager) generateHostEndpointManifest(endpoint *HostEndpoint) string {
// 	manifest := fmt.Sprintf(`apiVersion: crd.projectcalico.org/v1
// kind: HostEndpoint
// metadata:
//   name: %s
// spec:
//   node: %s
//   interfaceName: %s
// `, endpoint.Name, endpoint.Node, endpoint.Interface)
//
// 	if len(endpoint.ExpectedIPs) > 0 {
// 		manifest += "  expectedIPs:\n"
// 		for _, ip := range endpoint.ExpectedIPs {
// 			manifest += fmt.Sprintf("  - %s\n", ip)
// 		}
// 	}
//
// 	if len(endpoint.Profiles) > 0 {
// 		manifest += "  profiles:\n"
// 		for _, profile := range endpoint.Profiles {
// 			manifest += fmt.Sprintf("  - %s\n", profile)
// 		}
// 	}
//
// 	return manifest
// }
//
// // generateRuleYAML generates YAML for a network policy rule
// func (cm *CalicoManager) generateRuleYAML(rule Rule, indent string) string {
// 	var yaml strings.Builder
//
// 	yaml.WriteString(fmt.Sprintf("%s- action: %s\n", indent, rule.Action))
//
// 	if rule.Protocol != "" {
// 		yaml.WriteString(fmt.Sprintf("%s  protocol: %s\n", indent, rule.Protocol))
// 	}
//
// 	if len(rule.Source.Nets) > 0 {
// 		yaml.WriteString(fmt.Sprintf("%s  source:\n", indent))
// 		yaml.WriteString(fmt.Sprintf("%s    nets:\n", indent))
// 		for _, net := range rule.Source.Nets {
// 			yaml.WriteString(fmt.Sprintf("%s    - %s\n", indent, net))
// 		}
// 	}
//
// 	if len(rule.Source.NotNets) > 0 {
// 		yaml.WriteString(fmt.Sprintf("%s  source:\n", indent))
// 		yaml.WriteString(fmt.Sprintf("%s    notNets:\n", indent))
// 		for _, net := range rule.Source.NotNets {
// 			yaml.WriteString(fmt.Sprintf("%s    - %s\n", indent, net))
// 		}
// 	}
//
// 	if rule.Source.Selector != "" {
// 		yaml.WriteString(fmt.Sprintf("%s  source:\n", indent))
// 		yaml.WriteString(fmt.Sprintf("%s    selector: %s\n", indent, rule.Source.Selector))
// 	}
//
// 	if len(rule.Source.Ports) > 0 {
// 		yaml.WriteString(fmt.Sprintf("%s  source:\n", indent))
// 		yaml.WriteString(fmt.Sprintf("%s    ports:\n", indent))
// 		for _, port := range rule.Source.Ports {
// 			yaml.WriteString(fmt.Sprintf("%s    - port: %d\n", indent, port.Number))
// 			if port.Protocol != "" {
// 				yaml.WriteString(fmt.Sprintf("%s      protocol: %s\n", indent, port.Protocol))
// 			}
// 		}
// 	}
//
// 	if len(rule.Destination.Nets) > 0 {
// 		yaml.WriteString(fmt.Sprintf("%s  destination:\n", indent))
// 		yaml.WriteString(fmt.Sprintf("%s    nets:\n", indent))
// 		for _, net := range rule.Destination.Nets {
// 			yaml.WriteString(fmt.Sprintf("%s    - %s\n", indent, net))
// 		}
// 	}
//
// 	if len(rule.Destination.NotNets) > 0 {
// 		yaml.WriteString(fmt.Sprintf("%s  destination:\n", indent))
// 		yaml.WriteString(fmt.Sprintf("%s    notNets:\n", indent))
// 		for _, net := range rule.Destination.NotNets {
// 			yaml.WriteString(fmt.Sprintf("%s    - %s\n", indent, net))
// 		}
// 	}
//
// 	if rule.Destination.Selector != "" {
// 		yaml.WriteString(fmt.Sprintf("%s  destination:\n", indent))
// 		yaml.WriteString(fmt.Sprintf("%s    selector: %s\n", indent, rule.Destination.Selector))
// 	}
//
// 	if len(rule.Destination.Ports) > 0 {
// 		yaml.WriteString(fmt.Sprintf("%s  destination:\n", indent))
// 		yaml.WriteString(fmt.Sprintf("%s    ports:\n", indent))
// 		for _, port := range rule.Destination.Ports {
// 			yaml.WriteString(fmt.Sprintf("%s    - port: %d\n", indent, port.Number))
// 			if port.Protocol != "" {
// 				yaml.WriteString(fmt.Sprintf("%s      protocol: %s\n", indent, port.Protocol))
// 			}
// 		}
// 	}
//
// 	if rule.HTTP != nil {
// 		yaml.WriteString(fmt.Sprintf("%s  http:\n", indent))
// 		if len(rule.HTTP.Methods) > 0 {
// 			yaml.WriteString(fmt.Sprintf("%s    methods:\n", indent))
// 			for _, method := range rule.HTTP.Methods {
// 				yaml.WriteString(fmt.Sprintf("%s    - %s\n", indent, method))
// 			}
// 		}
// 		if len(rule.HTTP.Paths) > 0 {
// 			yaml.WriteString(fmt.Sprintf("%s    paths:\n", indent))
// 			for _, path := range rule.HTTP.Paths {
// 				yaml.WriteString(fmt.Sprintf("%s    - %s\n", indent, path))
// 			}
// 		}
// 	}
//
// 	if rule.ICMP != nil {
// 		yaml.WriteString(fmt.Sprintf("%s  icmp:\n", indent))
// 		yaml.WriteString(fmt.Sprintf("%s    type: %d\n", indent, rule.ICMP.Type))
// 		yaml.WriteString(fmt.Sprintf("%s    code: %d\n", indent, rule.ICMP.Code))
// 	}
//
// 	return yaml.String()
// }
