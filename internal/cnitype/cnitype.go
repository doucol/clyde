// Package cnitype provides functionality to detect the CNI (Container Network Interface) plugin used in a Kubernetes cluster.
package cnitype

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type CNIInfo struct {
	Name    string
	Version string
	Details map[string]string
}

// DetectCNI detects the CNI plugin being used in the Kubernetes cluster
func DetectCNI(kubeconfig string) (*CNIInfo, error) {
	// Create Kubernetes client
	client, err := createKubernetesClient(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ctx := context.Background()

	// Method 1: Check DaemonSets in kube-system namespace
	cniInfo, err := detectFromDaemonSets(ctx, client)
	if err == nil && cniInfo != nil {
		return cniInfo, nil
	}

	// Method 2: Check Deployments in kube-system namespace
	cniInfo, err = detectFromDeployments(ctx, client)
	if err == nil && cniInfo != nil {
		return cniInfo, nil
	}

	// Method 3: Check ConfigMaps for CNI configuration
	cniInfo, err = detectFromConfigMaps(ctx, client)
	if err == nil && cniInfo != nil {
		return cniInfo, nil
	}

	// Method 4: Check node annotations
	cniInfo, err = detectFromNodeAnnotations(ctx, client)
	if err == nil && cniInfo != nil {
		return cniInfo, nil
	}

	// Method 5: Detect cloud provider specific CNI
	cniInfo, err = detectCloudProviderCNI(ctx, client)
	if err == nil && cniInfo != nil {
		return cniInfo, nil
	}

	return nil, fmt.Errorf("unable to detect CNI plugin")
}

func createKubernetesClient(kubeconfig string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		// Use provided kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// Use in-cluster config (when running inside a pod)
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func detectFromDaemonSets(ctx context.Context, client *kubernetes.Clientset) (*CNIInfo, error) {
	daemonSets, err := client.AppsV1().DaemonSets("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	cniPatterns := map[string]string{
		"calico":  "calico",
		"flannel": "flannel",
		"weave":   "weave",
		"cilium":  "cilium",
		"antrea":  "antrea",
		"canal":   "canal",
		// AWS EKS CNI
		"aws-node":   "aws-vpc-cni",
		"kube-proxy": "aws-kube-proxy", // Often indicates EKS
		// Azure AKS CNI
		"azure-cni": "azure-cni",
		"azure-npm": "azure-npm", // Azure Network Policy Manager
		// GKE CNI
		"gke-node": "gke-native", // GKE native networking
		"ip-masq":  "gke-ip-masq-agent",
	}

	for _, ds := range daemonSets.Items {
		dsName := strings.ToLower(ds.Name)
		for pattern, cniName := range cniPatterns {
			if strings.Contains(dsName, pattern) {
				details := make(map[string]string)
				details["daemonset"] = ds.Name
				details["namespace"] = ds.Namespace

				// Try to get version from image tag
				version := "unknown"
				if len(ds.Spec.Template.Spec.Containers) > 0 {
					image := ds.Spec.Template.Spec.Containers[0].Image
					if parts := strings.Split(image, ":"); len(parts) > 1 {
						version = parts[1]
					}
				}

				return &CNIInfo{
					Name:    cniName,
					Version: version,
					Details: details,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("no CNI found in daemonsets")
}

func detectFromDeployments(ctx context.Context, client *kubernetes.Clientset) (*CNIInfo, error) {
	deployments, err := client.AppsV1().Deployments("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	cniPatterns := map[string]string{
		"calico":  "calico",
		"flannel": "flannel",
		"weave":   "weave",
		"cilium":  "cilium",
		"antrea":  "antrea",
		// Azure AKS specific
		"azure-cni": "azure-cni",
		"coredns":   "coredns", // Often customized in managed clusters
		// GKE specific
		"gke":        "gke-native",
		"l7-default": "gke-ingress",
	}

	for _, deploy := range deployments.Items {
		deployName := strings.ToLower(deploy.Name)
		for pattern, cniName := range cniPatterns {
			if strings.Contains(deployName, pattern) {
				details := make(map[string]string)
				details["deployment"] = deploy.Name
				details["namespace"] = deploy.Namespace

				version := "unknown"
				if len(deploy.Spec.Template.Spec.Containers) > 0 {
					image := deploy.Spec.Template.Spec.Containers[0].Image
					if parts := strings.Split(image, ":"); len(parts) > 1 {
						version = parts[1]
					}
				}

				return &CNIInfo{
					Name:    cniName,
					Version: version,
					Details: details,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("no CNI found in deployments")
}

func detectFromConfigMaps(ctx context.Context, client *kubernetes.Clientset) (*CNIInfo, error) {
	configMaps, err := client.CoreV1().ConfigMaps("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, cm := range configMaps.Items {
		cmName := strings.ToLower(cm.Name)

		// Check common CNI ConfigMap names
		if strings.Contains(cmName, "cni-config") || strings.Contains(cmName, "kube-proxy-config") {
			for key, data := range cm.Data {
				dataLower := strings.ToLower(data)

				cniTypes := map[string]string{
					"calico":  "calico",
					"flannel": "flannel",
					"weave":   "weave-net",
					"cilium":  "cilium",
					"antrea":  "antrea",
				}

				for cniType, cniName := range cniTypes {
					if strings.Contains(dataLower, cniType) {
						details := make(map[string]string)
						details["configmap"] = cm.Name
						details["key"] = key
						details["namespace"] = cm.Namespace

						return &CNIInfo{
							Name:    cniName,
							Version: "unknown",
							Details: details,
						}, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no CNI found in configmaps")
}

func detectFromNodeAnnotations(ctx context.Context, client *kubernetes.Clientset) (*CNIInfo, error) {
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		// Check common CNI annotations
		for key, value := range node.Annotations {
			keyLower := strings.ToLower(key)
			valueLower := strings.ToLower(value)

			if strings.Contains(keyLower, "cni") || strings.Contains(keyLower, "network") {
				cniTypes := map[string]string{
					"calico":  "calico",
					"flannel": "flannel",
					"weave":   "weave",
					"cilium":  "cilium",
					"antrea":  "antrea",
				}

				for cniType, cniName := range cniTypes {
					if strings.Contains(valueLower, cniType) {
						details := make(map[string]string)
						details["node"] = node.Name
						details["annotation"] = key
						details["value"] = value

						return &CNIInfo{
							Name:    cniName,
							Version: "unknown",
							Details: details,
						}, nil
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no CNI found in node annotations")
}

func detectCloudProviderCNI(ctx context.Context, client *kubernetes.Clientset) (*CNIInfo, error) {
	// Method 1: Check nodes for cloud provider labels and annotations
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) == 0 {
		return nil, fmt.Errorf("no nodes found")
	}

	node := nodes.Items[0] // Check first node for cloud provider info
	details := make(map[string]string)
	details["detection_method"] = "cloud_provider_analysis"

	// Check for EKS
	if isEKS(node) {
		// Verify AWS VPC CNI exists
		if hasAWSVPCCNI(ctx, client) {
			details["provider"] = "AWS EKS"
			details["node_instance_type"] = node.Labels["node.kubernetes.io/instance-type"]
			details["availability_zone"] = node.Labels["topology.kubernetes.io/zone"]

			return &CNIInfo{
				Name:    "aws-vpc-cni",
				Version: getAWSVPCCNIVersion(ctx, client),
				Details: details,
			}, nil
		}
	}

	// Check for AKS
	if isAKS(node) {
		cniType := getAKSCNIType(ctx, client, node)
		details["provider"] = "Azure AKS"
		details["aks_cluster"] = node.Labels["kubernetes.azure.com/cluster"]
		details["vm_size"] = node.Labels["node.kubernetes.io/instance-type"]

		return &CNIInfo{
			Name:    cniType,
			Version: getAKSCNIVersion(ctx, client, cniType),
			Details: details,
		}, nil
	}

	// Check for GKE
	if isGKE(node) {
		cniType := getGKECNIType(ctx, client, node)
		details["provider"] = "Google GKE"
		details["gke_nodepool"] = node.Labels["cloud.google.com/gke-nodepool"]
		details["machine_type"] = node.Labels["node.kubernetes.io/instance-type"]

		return &CNIInfo{
			Name:    cniType,
			Version: getGKECNIVersion(ctx, client),
			Details: details,
		}, nil
	}

	return nil, fmt.Errorf("no cloud provider CNI detected")
}

func isEKS(node corev1.Node) bool {
	// Check for EKS specific labels and annotations
	if provider, exists := node.Labels["kubernetes.io/cloud-provider"]; exists && provider == "aws" {
		return true
	}

	if _, exists := node.Labels["eks.amazonaws.com/nodegroup"]; exists {
		return true
	}

	if _, exists := node.Labels["alpha.eksctl.io/cluster-name"]; exists {
		return true
	}

	// Check node name pattern (EKS nodes often have specific naming)
	if strings.Contains(node.Name, "ip-") && strings.Contains(node.Name, ".ec2.internal") {
		return true
	}

	return false
}

func isAKS(node corev1.Node) bool {
	// Check for AKS specific labels
	if provider, exists := node.Labels["kubernetes.io/cloud-provider"]; exists && provider == "azure" {
		return true
	}

	if _, exists := node.Labels["kubernetes.azure.com/cluster"]; exists {
		return true
	}

	if _, exists := node.Labels["agentpool"]; exists {
		return true
	}

	// Check for AKS specific annotations
	for key := range node.Annotations {
		if strings.Contains(key, "azure") {
			return true
		}
	}

	return false
}

func isGKE(node corev1.Node) bool {
	// Check for GKE specific labels
	if provider, exists := node.Labels["kubernetes.io/cloud-provider"]; exists && provider == "gce" {
		return true
	}

	if _, exists := node.Labels["cloud.google.com/gke-nodepool"]; exists {
		return true
	}

	if _, exists := node.Labels["cloud.google.com/gke-os-distribution"]; exists {
		return true
	}

	return false
}

func hasAWSVPCCNI(ctx context.Context, client *kubernetes.Clientset) bool {
	// Check for AWS VPC CNI DaemonSet
	_, err := client.AppsV1().DaemonSets("kube-system").Get(ctx, "aws-node", metav1.GetOptions{})
	return err == nil
}

func getAWSVPCCNIVersion(ctx context.Context, client *kubernetes.Clientset) string {
	ds, err := client.AppsV1().DaemonSets("kube-system").Get(ctx, "aws-node", metav1.GetOptions{})
	if err != nil {
		return "unknown"
	}

	for _, container := range ds.Spec.Template.Spec.Containers {
		if container.Name == "aws-node" {
			if parts := strings.Split(container.Image, ":"); len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return "unknown"
}

func getAKSCNIType(ctx context.Context, client *kubernetes.Clientset, node corev1.Node) string {
	// Check for Azure CNI vs Kubenet
	if networkPlugin, exists := node.Labels["kubernetes.azure.com/network-plugin"]; exists {
		if networkPlugin == "azure" {
			return "azure-cni"
		} else if networkPlugin == "kubenet" {
			return "kubenet"
		}
	}

	// Check for Azure NPM (Network Policy Manager)
	if _, err := client.AppsV1().DaemonSets("kube-system").Get(ctx, "azure-npm", metav1.GetOptions{}); err == nil {
		return "azure-cni-with-npm"
	}

	// Default to Azure CNI for AKS
	return "azure-cni"
}

func getAKSCNIVersion(ctx context.Context, client *kubernetes.Clientset, cniType string) string {
	// Try to get version from azure-cni-networkmonitor pod
	pods, err := client.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "component=azure-cni-networkmonitor",
	})
	if err == nil && len(pods.Items) > 0 {
		for _, container := range pods.Items[0].Spec.Containers {
			if strings.Contains(container.Image, "azure") {
				if parts := strings.Split(container.Image, ":"); len(parts) > 1 {
					return parts[1]
				}
			}
		}
	}
	return "unknown"
}

func getGKECNIType(ctx context.Context, client *kubernetes.Clientset, node corev1.Node) string {
	// Check for Dataplane V2 (Cilium)
	if _, err := client.AppsV1().DaemonSets("kube-system").Get(ctx, "cilium", metav1.GetOptions{}); err == nil {
		return "gke-dataplane-v2-cilium"
	}

	// Check for network policy
	if networkPolicy, exists := node.Labels["cloud.google.com/gke-network-policy"]; exists {
		if networkPolicy == "calico" {
			return "gke-native-with-calico"
		}
	}

	// Check cluster networking mode from node labels
	if mode, exists := node.Labels["cloud.google.com/gke-networking-mode"]; exists {
		return fmt.Sprintf("gke-%s", mode)
	}

	return "gke-native"
}

func getGKECNIVersion(ctx context.Context, client *kubernetes.Clientset) string {
	// Try to get GKE version from node
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil || len(nodes.Items) == 0 {
		return "unknown"
	}

	node := nodes.Items[0]
	if version, exists := node.Labels["cloud.google.com/gke-version"]; exists {
		return version
	}

	return node.Status.NodeInfo.KubeletVersion
}
