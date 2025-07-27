package util

import (
	context "context"
	"errors"
	fmt "fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/doucol/clyde/internal/githubversions"
	"github.com/doucol/clyde/internal/k8sapply"
	"golang.org/x/mod/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ClusterNetworkingInfo holds metadata about the cluster's networking
// and Calico/operator status
type ClusterNetworkingInfo struct {
	CNIType           string
	PodCIDRs          []string
	ServiceCIDRs      []string
	Overlay           string
	Encapsulation     string
	CalicoInstalled   bool
	CalicoVersion     string
	OperatorInstalled bool
	OperatorVersion   string
	WhiskerAvailable  bool
	CalicoNamespace   string
	CalicoOperatorNS  string
	CalicoPods        []string
	OperatorPods      []string
	Errors            []string
}

// DetectCNIType inspects kube-system and calico-system DaemonSets/pods for known CNI plugins
func DetectCNIType(ctx context.Context, clientset kubernetes.Interface) (string, error) {
	// Check multiple namespaces where CNI components might be deployed
	namespaces := []string{"kube-system", "calico-system", "tigera-operator"}

	for _, namespace := range namespaces {
		// Check DaemonSets first
		daemonsets, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, ds := range daemonsets.Items {
				name := strings.ToLower(ds.Name)
				if strings.Contains(name, "calico") || name == "calico-node" {
					return "Calico", nil
				}
				if strings.Contains(name, "cilium") {
					return "Cilium", nil
				}
				if strings.Contains(name, "weave") {
					return "Weave", nil
				}
				if strings.Contains(name, "flannel") {
					return "Flannel", nil
				}
				if strings.Contains(name, "canal") {
					return "Canal", nil
				}
				if strings.Contains(name, "kube-router") {
					return "Kube-Router", nil
				}
				if strings.Contains(name, "ovn") {
					return "OVN-Kubernetes", nil
				}
			}
		}

		// Also check pods for Calico images (especially useful for operator-managed installs)
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, pod := range pods.Items {
				for _, container := range pod.Spec.Containers {
					image := strings.ToLower(container.Image)
					if strings.Contains(image, "calico/node") || strings.Contains(image, "calico/cni") {
						return "Calico", nil
					}
					if strings.Contains(image, "cilium") {
						return "Cilium", nil
					}
					if strings.Contains(image, "weave") {
						return "Weave", nil
					}
					if strings.Contains(image, "flannel") {
						return "Flannel", nil
					}
				}
			}
		}
	}

	return "Unknown", nil
}

// GetCalicoPods returns pod names and versions for Calico in calico-system
func GetCalicoPods(ctx context.Context, clientset kubernetes.Interface, ns string) (pods []string, version string, err error) {
	podList, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, "", err
	}
	verSet := map[string]struct{}{}
	for _, pod := range podList.Items {
		pods = append(pods, pod.Name)
		for _, c := range pod.Spec.Containers {
			if strings.Contains(c.Image, "calico") {
				ver := parseImageVersion(c.Image)
				if ver != "" {
					verSet[ver] = struct{}{}
				}
			}
		}
	}
	if len(verSet) > 0 {
		vers := make([]string, 0, len(verSet))
		for v := range verSet {
			vers = append(vers, v)
		}
		sort.Strings(vers)
		version = vers[len(vers)-1]
	}
	return pods, version, nil
}

// GetOperatorPods returns pod names and versions for Calico operator
func GetOperatorPods(ctx context.Context, clientset kubernetes.Interface, ns string) (pods []string, version string, err error) {
	if ns == "" {
		ns = "calico-system"
	}

	// Try multiple namespaces where the operator might be installed
	namespaces := []string{ns, "tigera-operator", "calico-operator"}

	verSet := map[string]struct{}{}

	for _, namespace := range namespaces {
		// Try different label selectors that are commonly used for Tigera/Calico operator
		labelSelectors := []string{
			"k8s-app=tigera-operator",
			"name=tigera-operator",
			"app.kubernetes.io/name=tigera-operator",
			"k8s-app=calico-operator",
			"name=calico-operator",
		}

		for _, selector := range labelSelectors {
			podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
			if err != nil {
				continue // Try next selector
			}

			for _, pod := range podList.Items {
				pods = append(pods, pod.Name)
				for _, c := range pod.Spec.Containers {
					if strings.Contains(c.Image, "tigera") || strings.Contains(c.Image, "operator") || strings.Contains(c.Image, "calico") {
						ver := parseImageVersion(c.Image)
						if ver != "" {
							verSet[ver] = struct{}{}
						}
					}
				}
			}
		}

		// Also try looking for deployments with operator in the name
		deployments, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, deploy := range deployments.Items {
				deployName := strings.ToLower(deploy.Name)
				if strings.Contains(deployName, "operator") || strings.Contains(deployName, "tigera") {
					pods = append(pods, deploy.Name+"-deployment")
					for _, c := range deploy.Spec.Template.Spec.Containers {
						if strings.Contains(c.Image, "tigera") || strings.Contains(c.Image, "operator") || strings.Contains(c.Image, "calico") {
							ver := parseImageVersion(c.Image)
							if ver != "" {
								verSet[ver] = struct{}{}
							}
						}
					}
				}
			}
		}

		// If we found something in this namespace, we can break
		if len(pods) > 0 {
			break
		}
	}

	if len(verSet) > 0 {
		vers := make([]string, 0, len(verSet))
		for v := range verSet {
			vers = append(vers, v)
		}
		sort.Strings(vers)
		version = vers[len(vers)-1]
	}
	return pods, version, nil
}

// parseImageVersion extracts the version from an image string (e.g. calico/node:v3.30.0)
func parseImageVersion(image string) string {
	parts := strings.Split(image, ":")
	if len(parts) == 2 {
		return strings.TrimPrefix(parts[1], "v")
	}
	return ""
}

// CompareVersions returns true if v1 >= v2 (semantic versioning)
func CompareVersions(v1, v2 string) bool {
	if !strings.HasPrefix(v1, "v") {
		v1 = "v" + v1
	}
	if !strings.HasPrefix(v2, "v") {
		v2 = "v" + v2
	}
	v1Valid := semver.IsValid(v1)
	v2Valid := semver.IsValid(v2)
	if v1Valid && v2Valid {
		return semver.Compare(v1, v2) >= 0
	} else if v1Valid && !v2Valid {
		return true
	}
	return false
}

func InstallCalicoOperator(ctx context.Context, clientset kubernetes.Interface, dyn dynamic.Interface) error {
	var err error
	var latest string
	var applier *k8sapply.Applier
	if latest, err = githubversions.GetLatestStableSemverTag(ctx, "projectcalico", "calico"); err != nil || len(latest) == 0 {
		if err != nil {
			return err
		}
		return errors.New("failed to fetch latest Calico version - calico version is empty")
	}
	applier, err = k8sapply.NewApplier(clientset, dyn, nil)
	if err != nil {
		return fmt.Errorf("failed to create k8sapplier: %w", err)
	}
	err = applier.ApplyURL(ctx, fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%s/manifests/tigera-operator.yaml", latest))
	if err != nil {
		return fmt.Errorf("failed to install Calico %s: %w", latest, err)
	}
	err = applier.ApplyURL(ctx, fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%s/manifests/custom-resources.yaml", latest))
	if err != nil {
		return fmt.Errorf("failed to install Calico %s custom resources: %w", latest, err)
	}
	return nil
}

// GetPodCIDRs tries to fetch pod CIDRs from kubeadm-config ConfigMap and/or Calico IPPool CRDs
func GetPodCIDRs(ctx context.Context, clientset kubernetes.Interface, dyn dynamic.Interface) ([]string, error) {
	var cidrs []string
	// Try kubeadm-config ConfigMap
	cm, err := clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "kubeadm-config", metav1.GetOptions{})
	if err == nil && cm.Data != nil {
		if clusterConfig, ok := cm.Data["ClusterConfiguration"]; ok {
			// Look for podSubnet: ...
			re := regexp.MustCompile(`podSubnet:\s*([\d./,:a-fA-F]+)`)
			if m := re.FindStringSubmatch(clusterConfig); len(m) > 1 {
				cidrs = append(cidrs, strings.Split(m[1], ",")...)
			}
		}
	}
	// Try Calico IPPool CRDs (if available)
	if dyn != nil {
		gvr := schema.GroupVersionResource{Group: "crd.projectcalico.org", Version: "v1", Resource: "ippools"}
		ippools, err := dyn.Resource(gvr).List(ctx, metav1.ListOptions{})
		if err == nil {
			for _, item := range ippools.Items {
				if spec, ok := item.Object["spec"].(map[string]interface{}); ok {
					if cidr, ok := spec["cidr"].(string); ok {
						cidrs = append(cidrs, cidr)
					}
				}
			}
		}
	}
	return cidrs, nil
}

// GetServiceCIDRs tries to fetch service CIDRs from kubeadm-config ConfigMap
func GetServiceCIDRs(ctx context.Context, clientset kubernetes.Interface) ([]string, error) {
	var cidrs []string
	cm, err := clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "kubeadm-config", metav1.GetOptions{})
	if err == nil && cm.Data != nil {
		if clusterConfig, ok := cm.Data["ClusterConfiguration"]; ok {
			re := regexp.MustCompile(`serviceSubnet:\s*([\d./,:a-fA-F]+)`)
			if m := re.FindStringSubmatch(clusterConfig); len(m) > 1 {
				cidrs = append(cidrs, strings.Split(m[1], ",")...)
			}
		}
	}
	return cidrs, nil
}

// GetOverlayAndEncapsulation tries to fetch overlay/encapsulation from Calico IPPool CRDs
func GetOverlayAndEncapsulation(ctx context.Context, dyn dynamic.Interface) (overlay, encaps string, err error) {
	if dyn == nil {
		return "", "", nil
	}
	gvr := schema.GroupVersionResource{Group: "crd.projectcalico.org", Version: "v1", Resource: "ippools"}
	ippools, err := dyn.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", "", nil
	}
	for _, item := range ippools.Items {
		if spec, ok := item.Object["spec"].(map[string]interface{}); ok {
			if vxlan, ok := spec["vxlanMode"].(string); ok && vxlan != "Never" {
				overlay = "VXLAN"
				encaps = vxlan
			}
			if ipip, ok := spec["ipipMode"].(string); ok && ipip != "Never" {
				overlay = "IPIP"
				encaps = ipip
			}
			if nat, ok := spec["natOutgoing"].(bool); ok && nat {
				if overlay == "" {
					overlay = "NAT"
				}
			}
		}
	}
	return overlay, encaps, nil
}

// GetWhiskerAvailability checks if whisker-backend container is available in calico-system
func GetWhiskerAvailability(ctx context.Context, clientset kubernetes.Interface, namespace string) bool {
	if namespace == "" {
		namespace = "calico-system"
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false
	}

	for _, pod := range pods.Items {
		if pod.Status.Phase != "Running" {
			continue
		}
		for _, container := range pod.Spec.Containers {
			if container.Name == "whisker-backend" {
				return true
			}
		}
	}
	return false
}

// GetClusterNetworkingInfo gathers all cluster networking information
func GetClusterNetworkingInfo(ctx context.Context, clientset kubernetes.Interface, dyn dynamic.Interface, restConfig *rest.Config) ClusterNetworkingInfo {
	info := ClusterNetworkingInfo{
		CalicoNamespace:  "calico-system",
		CalicoOperatorNS: "calico-system",
	}
	var errs []string
	cni, err := DetectCNIType(ctx, clientset)
	if err != nil {
		errs = append(errs, "CNI detection: "+err.Error())
	}
	info.CNIType = cni
	pods, ver, err := GetCalicoPods(ctx, clientset, info.CalicoNamespace)
	if err == nil && len(pods) > 0 {
		info.CalicoInstalled = true
		info.CalicoVersion = ver
		info.CalicoPods = pods
	}
	pods, ver, err = GetOperatorPods(ctx, clientset, info.CalicoOperatorNS)
	if err == nil && len(pods) > 0 {
		info.OperatorInstalled = true
		info.OperatorVersion = ver
		info.OperatorPods = pods
	}
	podCIDRs, err := GetPodCIDRs(ctx, clientset, dyn)
	if err == nil && len(podCIDRs) > 0 {
		info.PodCIDRs = podCIDRs
	}
	serviceCIDRs, err := GetServiceCIDRs(ctx, clientset)
	if err == nil && len(serviceCIDRs) > 0 {
		info.ServiceCIDRs = serviceCIDRs
	}
	overlay, encaps, err := GetOverlayAndEncapsulation(ctx, dyn)
	if err == nil {
		info.Overlay = overlay
		info.Encapsulation = encaps
	}
	// Check whisker availability
	info.WhiskerAvailable = GetWhiskerAvailability(ctx, clientset, info.CalicoNamespace)

	info.Errors = errs
	return info
}
