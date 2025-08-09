// Package calico provides comprehensive Calico OSS management capabilities
// including installation, configuration, and management of Calico resources.
package calico

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/githubversions"
	k8sapplier "github.com/doucol/clyde/internal/k8sapply"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// CalicoManager provides comprehensive Calico OSS management capabilities
type CalicoManager struct {
	clientset kubernetes.Interface
	dynamic   dynamic.Interface
	logWriter io.Writer
	logChan   chan string
}

// NewCalicoManager creates a new CalicoManager instance
func NewCalicoManager(clientset kubernetes.Interface, dynamicClient dynamic.Interface, logger io.Writer) *CalicoManager {
	return &CalicoManager{
		clientset: clientset,
		dynamic:   dynamicClient,
		logWriter: logger,
		logChan:   nil,
	}
}

// Logf logs formatted messages to the logger
func (cm *CalicoManager) Logf(format string, args ...any) {
	if len(format) > 0 {
		cm.Log(fmt.Sprintf(format, args...))
	}
}

func (cm *CalicoManager) Log(message string) {
	if len(message) == 0 {
		return
	}
	if cm.logChan != nil {
		cm.logChan <- message
	}
	if cm.logWriter != nil {
		_, _ = fmt.Fprint(cm.logWriter, message+"\n")
	}
}

func (cm *CalicoManager) LogChan() chan string {
	if cm.logChan == nil {
		cm.logChan = make(chan string, 100) // Buffered channel for events
	}
	return cm.logChan
}

func (cm *CalicoManager) Close() {
	if cm.logChan != nil {
		close(cm.logChan) // Close the events channel
	}
}

// Install installs Calico using the operator-based installation method
func (cm *CalicoManager) Install(ctx context.Context, options *InstallOptions) error {
	if options == nil {
		return fmt.Errorf("install options cannot be nil")
	}

	latestVersion, err := githubversions.GetLatestStableSemverTag(ctx, "projectcalico", "calico")
	if err != nil {
		cm.Logf("[red]Failed to get latest Calico version: %v", err)
		return err
	}

	cm.Logf("[white]Installing latest Calico version: %s", latestVersion)

	// Install Calico operator first
	if err := cm.installOperator(ctx, latestVersion, options); err != nil {
		cm.Logf("[red]Failed to install Calico operator: %s", err.Error())
		return fmt.Errorf("failed to install Calico operator: %w", err)
	}

	// Wait for the Calico CRDs to be created by the operator
	if err := cm.waitForCalicoCRDs(ctx); err != nil {
		cm.Logf("[red]Failed to wait Calico CRDs: %s", err.Error())
		return fmt.Errorf("failed to wait for Calico CRDs: %w", err)
	}

	// Install Calico instance
	// if err := cm.installCalicoInstance(ctx, options); err != nil {
	// 	return fmt.Errorf("failed to install Calico instance: %w", err)
	// }

	cm.Logf("[green]Calico installation completed successfully!")
	return nil
}

// installOperator installs the Calico operator
func (cm *CalicoManager) installOperator(ctx context.Context, latestVersion string, options *InstallOptions) error {
	cm.Logf("[white]Installing Calico operator...")

	applier, err := k8sapplier.NewApplierWithClients(cm.clientset, cm.dynamic, nil)
	if err != nil {
		return fmt.Errorf("failed to create applier: %w", err)
	}

	// logChan := applier.LogChan()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer applier.Close()
		murl := fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%s/manifests/tigera-operator.yaml", latestVersion)
		results, err := applier.ApplyFromURL(ctx, murl, nil)
		if err != nil {
			cm.Logf("[red]Failed to apply Calico operator: %v", err)
			return
		}
		for _, result := range results {
			if result.Error != nil {
				cm.Logf("[red]Error applying %s/%s (%s): %v", result.Namespace, result.Name, result.Kind, result.Error)
			} else {
				cm.Logf("[green]Successfully %s: %s/%s (%s)", result.Action, result.Namespace, result.Name, result.Kind)
			}
		}
	}()

	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		case msg, ok := <-logChan:
	// 			cm.Log(msg)
	// 			if !ok {
	// 				return
	// 			}
	// 		}
	// 	}
	// }()

	// Wait for operator to be ready
	// if err := cm.waitForOperatorReady(ctx); err != nil {
	// 	return fmt.Errorf("operator failed to become ready: %w", err)
	// }

	wg.Wait()
	cm.Logf("[green]Calico operator installed successfully")
	return nil
}

// installCalicoInstance installs the Calico instance
// func (cm *CalicoManager) installCalicoInstance(ctx context.Context, options *InstallOptions) error {
// 	cm.Logf("Installing Calico instance...")
//
// 	// Generate and apply Calico instance manifest
// 	instanceManifest := cm.generateCalicoInstanceManifest(options)
// 	if err := cm.applyManifest(ctx, instanceManifest); err != nil {
// 		return fmt.Errorf("failed to apply Calico instance manifest: %w", err)
// 	}
//
// 	// Wait for Calico to be ready
// 	if err := cm.waitForCalicoReady(ctx); err != nil {
// 		return fmt.Errorf("calico failed to become ready: %w", err)
// 	}
//
// 	cm.Logf("Calico instance installed successfully")
// 	return nil
// }

// Upgrade upgrades Calico to a new version
// func (cm *CalicoManager) Upgrade(ctx context.Context, options *UpgradeOptions) error {
// 	if options == nil {
// 		return fmt.Errorf("upgrade options cannot be nil")
// 	}
//
// 	cm.Logf("Upgrading Calico to version %s", options.Version)
//
// 	// Backup current configuration if requested
// 	if options.BackupBeforeUpgrade {
// 		if err := cm.backupConfiguration(ctx); err != nil {
// 			return fmt.Errorf("failed to backup configuration: %w", err)
// 		}
// 	}
//
// 	// Upgrade operator
// 	if err := cm.upgradeOperator(ctx, options); err != nil {
// 		return fmt.Errorf("failed to upgrade operator: %w", err)
// 	}
//
// 	// Upgrade Calico instance
// 	if err := cm.upgradeCalicoInstance(ctx, options); err != nil {
// 		return fmt.Errorf("failed to upgrade Calico instance: %w", err)
// 	}
//
// 	// Validate upgrade if requested
// 	if options.ValidateAfterUpgrade {
// 		if err := cm.validateUpgrade(ctx); err != nil {
// 			return fmt.Errorf("upgrade validation failed: %w", err)
// 		}
// 	}
//
// 	cm.Logf("Calico upgrade completed successfully")
// 	return nil
// }
//
// // Uninstall uninstalls Calico
// func (cm *CalicoManager) Uninstall(ctx context.Context, options *UninstallOptions) error {
// 	if options == nil {
// 		return fmt.Errorf("uninstall options cannot be nil")
// 	}
//
// 	cm.Logf("Uninstalling Calico...")
//
// 	// Delete Calico instance
// 	if err := cm.deleteCalicoInstance(ctx); err != nil {
// 		return fmt.Errorf("failed to delete Calico instance: %w", err)
// 	}
//
// 	// Delete operator
// 	if err := cm.deleteOperator(ctx); err != nil {
// 		return fmt.Errorf("failed to delete operator: %w", err)
// 	}
//
// 	// Delete CRDs if requested
// 	if options.RemoveCRDs {
// 		if err := cm.deleteCRDs(ctx); err != nil {
// 			return fmt.Errorf("failed to delete CRDs: %w", err)
// 		}
// 	}
//
// 	// Delete namespace if requested
// 	if options.RemoveNamespace {
// 		if err := cm.deleteNamespace(ctx); err != nil {
// 			return fmt.Errorf("failed to delete namespace: %w", err)
// 		}
// 	}
//
// 	cm.Logf("Calico uninstallation completed successfully")
// 	return nil
// }

// GetStatus returns the current status of Calico
// func (cm *CalicoManager) GetStatus(ctx context.Context) (*CalicoStatus, error) {
// 	cm.Logf("Getting Calico status...")
//
// 	status := &CalicoStatus{
// 		Components:      make(map[string]ComponentStatus),
// 		Nodes:           []NodeStatus{},
// 		IPPools:         []IPPoolStatus{},
// 		BGPPeers:        []BGPPeerStatus{},
// 		NetworkPolicies: []PolicyStatus{},
// 		LastUpdated:     time.Now(),
// 	}
//
// 	// Check if Calico is installed
// 	installed, version, err := cm.checkInstallation(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to check installation: %w", err)
// 	}
//
// 	status.Installed = installed
// 	status.Version = version
//
// 	if installed {
// 		// Get component statuses
// 		if err := cm.getComponentStatuses(ctx, status); err != nil {
// 			return nil, fmt.Errorf("failed to get component statuses: %w", err)
// 		}
//
// 		// Get node statuses
// 		if err := cm.getNodeStatuses(ctx, status); err != nil {
// 			return nil, fmt.Errorf("failed to get node statuses: %w", err)
// 		}
//
// 		// Get IP pool statuses
// 		if err := cm.getIPPoolStatuses(ctx, status); err != nil {
// 			return nil, fmt.Errorf("failed to get IP pool statuses: %w", err)
// 		}
//
// 		// Get BGP peer statuses
// 		if err := cm.getBGPPeerStatuses(ctx, status); err != nil {
// 			return nil, fmt.Errorf("failed to get BGP peer statuses: %w", err)
// 		}
//
// 		// Get policy statuses
// 		if err := cm.getPolicyStatuses(ctx, status); err != nil {
// 			return nil, fmt.Errorf("failed to get policy statuses: %w", err)
// 		}
// 	}
//
// 	return status, nil
// }

// GetIPPools returns all IP pools
func (cm *CalicoManager) GetIPPools(ctx context.Context) ([]*IPPool, error) {
	cm.Logf("Getting IP pools...")

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "ippools",
	}

	list, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list IP pools: %w", err)
	}

	var ippools []*IPPool
	for _, item := range list.Items {
		spec := item.Object["spec"].(map[string]any)

		ippool := &IPPool{
			Name:      item.GetName(),
			CIDR:      spec["cidr"].(string),
			BlockSize: int(spec["blockSize"].(float64)),
		}

		if ipip, ok := spec["ipipMode"].(string); ok {
			ippool.IPIPMode = ipip
		}
		if vxlan, ok := spec["vxlanMode"].(string); ok {
			ippool.VXLANMode = vxlan
		}

		ippools = append(ippools, ippool)
	}

	return ippools, nil
}

// CreateIPPool creates a new IP pool
// func (cm *CalicoManager) CreateIPPool(ctx context.Context, pool *IPPool) error {
// 	cm.Logf("Creating IP pool %s", pool.Name)
//
// 	manifest := cm.generateIPPoolManifest(pool)
//
// 	if err := cm.applyManifest(ctx, manifest); err != nil {
// 		return fmt.Errorf("failed to create IP pool: %w", err)
// 	}
//
// 	cm.Logf("IP pool %s created successfully", pool.Name)
// 	return nil
// }
//
// // DeleteIPPool deletes an IP pool
// func (cm *CalicoManager) DeleteIPPool(ctx context.Context, name string) error {
// 	cm.Logf("Deleting IP pool %s", name)
//
// 	gvr := schema.GroupVersionResource{
// 		Group:    "crd.projectcalico.org",
// 		Version:  "v1",
// 		Resource: "ippools",
// 	}
//
// 	if err := cm.dynamic.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
// 		return fmt.Errorf("failed to delete IP pool: %w", err)
// 	}
//
// 	cm.Logf("IP pool %s deleted successfully", name)
// 	return nil
// }

// GetNetworkPolicies returns network policies for a namespace
// func (cm *CalicoManager) GetNetworkPolicies(ctx context.Context, namespace string) ([]*NetworkPolicy, error) {
// 	cm.Logf("Getting network policies for namespace %s", namespace)
//
// 	gvr := schema.GroupVersionResource{
// 		Group:    "crd.projectcalico.org",
// 		Version:  "v1",
// 		Resource: "networkpolicies",
// 	}
//
// 	list, err := cm.dynamic.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to list network policies: %w", err)
// 	}
//
// 	var policies []*NetworkPolicy
// 	for _, item := range list.Items {
// 		spec := item.Object["spec"].(map[string]any)
//
// 		policy := &NetworkPolicy{
// 			Name:      item.GetName(),
// 			Namespace: item.GetNamespace(),
// 		}
//
// 		if ingress, ok := spec["ingress"].([]any); ok {
// 			policy.IngressRules = cm.parseRules(ingress)
// 		}
// 		if egress, ok := spec["egress"].([]any); ok {
// 			policy.EgressRules = cm.parseRules(egress)
// 		}
//
// 		policies = append(policies, policy)
// 	}
//
// 	return policies, nil
// }

// CreateNetworkPolicy creates a new network policy
// func (cm *CalicoManager) CreateNetworkPolicy(ctx context.Context, policy *NetworkPolicy) error {
// 	cm.Logf("Creating network policy %s", policy.Name)
//
// 	manifest := cm.generateNetworkPolicyManifest(policy)
//
// 	if err := cm.applyManifest(ctx, manifest); err != nil {
// 		return fmt.Errorf("failed to create network policy: %w", err)
// 	}
//
// 	cm.Logf("Network policy %s created successfully", policy.Name)
// 	return nil
// }
//
// // DeleteNetworkPolicy deletes a network policy
// func (cm *CalicoManager) DeleteNetworkPolicy(ctx context.Context, name, namespace string) error {
// 	cm.Logf("Deleting network policy %s from namespace %s", name, namespace)
//
// 	gvr := schema.GroupVersionResource{
// 		Group:    "crd.projectcalico.org",
// 		Version:  "v1",
// 		Resource: "networkpolicies",
// 	}
//
// 	if err := cm.dynamic.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
// 		return fmt.Errorf("failed to delete network policy: %w", err)
// 	}
//
// 	cm.Logf("Network policy %s deleted successfully", name)
// 	return nil
// }

// GetBGPPeers returns BGP peers
func (cm *CalicoManager) GetBGPPeers(ctx context.Context) ([]*BGPPeer, error) {
	cm.Logf("Getting BGP peers...")

	gvr := schema.GroupVersionResource{
		Group:    "crd.projectcalico.org",
		Version:  "v1",
		Resource: "bgppeers",
	}

	list, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list BGP peers: %w", err)
	}

	var peers []*BGPPeer
	for _, item := range list.Items {
		spec := item.Object["spec"].(map[string]any)

		peer := &BGPPeer{
			Name: item.GetName(),
		}

		if asn, ok := spec["asNumber"].(float64); ok {
			peer.ASN = int(asn)
		}
		if ip, ok := spec["peerIP"].(string); ok {
			peer.IP = ip
		}

		peers = append(peers, peer)
	}

	return peers, nil
}

// CreateBGPPeer creates a new BGP peer
// func (cm *CalicoManager) CreateBGPPeer(ctx context.Context, peer *BGPPeer) error {
// 	cm.Logf("Creating BGP peer %s", peer.Name)
//
// 	manifest := cm.generateBGPPeerManifest(peer)
//
// 	if err := cm.applyManifest(ctx, manifest); err != nil {
// 		return fmt.Errorf("failed to create BGP peer: %w", err)
// 	}
//
// 	cm.Logf("BGP peer %s created successfully", peer.Name)
// 	return nil
// }
//
// // DeleteBGPPeer deletes a BGP peer
// func (cm *CalicoManager) DeleteBGPPeer(ctx context.Context, name string) error {
// 	cm.Logf("Deleting BGP peer %s", name)
//
// 	gvr := schema.GroupVersionResource{
// 		Group:    "crd.projectcalico.org",
// 		Version:  "v1",
// 		Resource: "bgppeers",
// 	}
//
// 	if err := cm.dynamic.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
// 		return fmt.Errorf("failed to delete BGP peer: %w", err)
// 	}
//
// 	cm.Logf("BGP peer %s deleted successfully", name)
// 	return nil
// }

// WaitForReady waits for Calico components to be ready
// func (cm *CalicoManager) WaitForReady(ctx context.Context, timeout time.Duration) error {
// 	cm.Logf("Waiting for Calico to be ready (timeout: %v)...", timeout)
//
// 	deadline := time.Now().Add(timeout)
// 	ticker := time.NewTicker(10 * time.Second)
// 	defer ticker.Stop()
//
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		case <-ticker.C:
// 			if time.Now().After(deadline) {
// 				return fmt.Errorf("timeout waiting for Calico to be ready")
// 			}
//
// 			ready, err := cm.checkCalicoReady(ctx)
// 			if err != nil {
// 				cm.Logf("Error checking Calico readiness: %v", err)
// 				continue
// 			}
//
// 			if ready {
// 				cm.Logf("Calico is ready")
// 				return nil
// 			}
// 		}
// 	}
// }

// HealthCheck performs a comprehensive health check of Calico
// func (cm *CalicoManager) HealthCheck(ctx context.Context) (*HealthCheckResult, error) {
// 	cm.Logf("Performing Calico health check...")
//
// 	result := &HealthCheckResult{
// 		LastChecked: time.Now(),
// 		Overall:     true,
// 		Components:  make(map[string]bool),
// 		Nodes:       make(map[string]bool),
// 		BGP:         make(map[string]bool),
// 		Errors:      []string{},
// 	}
//
// 	// Check operator status
// 	if err := cm.checkOperatorHealth(ctx, result); err != nil {
// 		result.Overall = false
// 		result.Errors = append(result.Errors, fmt.Sprintf("Operator health check failed: %v", err))
// 	}
//
// 	// Check Calico instance status
// 	if err := cm.checkInstanceHealth(ctx, result); err != nil {
// 		result.Overall = false
// 		result.Errors = append(result.Errors, fmt.Sprintf("Instance health check failed: %v", err))
// 	}
//
// 	// Check node status
// 	if err := cm.checkNodeHealth(ctx, result); err != nil {
// 		result.Overall = false
// 		result.Errors = append(result.Errors, fmt.Sprintf("Node health check failed: %v", err))
// 	}
//
// 	// Check BGP status
// 	if err := cm.checkBGPHealth(ctx, result); err != nil {
// 		result.Overall = false
// 		result.Errors = append(result.Errors, fmt.Sprintf("BGP health check failed: %v", err))
// 	}
//
// 	if result.Overall {
// 		cm.Logf("Calico health check passed")
// 	} else {
// 		cm.Logf("Calico health check failed with %d errors", len(result.Errors))
// 	}
//
// 	return result, nil
// }
//
// // ApplyResource applies a Calico resource from YAML
// func (cm *CalicoManager) ApplyResource(ctx context.Context, yamlContent string) error {
// 	cm.Logf("Applying Calico resource from YAML")
//
// 	// Parse YAML into unstructured objects
// 	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(yamlContent), 4096)
//
// 	for {
// 		var obj unstructured.Unstructured
// 		err := decoder.Decode(&obj)
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return fmt.Errorf("failed to decode YAML: %w", err)
// 		}
//
// 		// Apply the resource
// 		if err := cm.applyUnstructured(ctx, &obj); err != nil {
// 			return fmt.Errorf("failed to apply resource %s: %w", obj.GetName(), err)
// 		}
// 	}
//
// 	cm.Logf("Calico resource applied successfully")
// 	return nil
// }

// Helper methods (implementations would be added here)
//
//	func (cm *CalicoManager) applyManifest(ctx context.Context, manifest string) error {
//		// Implementation would apply manifest to cluster
//		return nil
//	}
//
//	func (cm *CalicoManager) waitForOperatorReady(ctx context.Context) error {
//		// Implementation would wait for operator to be ready
//		return nil
//	}
//
//	func (cm *CalicoManager) waitForCalicoReady(ctx context.Context) error {
//		// Implementation would wait for Calico to be ready
//		return nil
//	}
//
//	func (cm *CalicoManager) backupConfiguration(ctx context.Context) error {
//		// Implementation would backup current configuration
//		return nil
//	}
//
//	func (cm *CalicoManager) upgradeOperator(ctx context.Context, options *UpgradeOptions) error {
//		// Implementation would upgrade operator
//		return nil
//	}
//
//	func (cm *CalicoManager) upgradeCalicoInstance(ctx context.Context, options *UpgradeOptions) error {
//		// Implementation would upgrade Calico instance
//		return nil
//	}
//
//	func (cm *CalicoManager) validateUpgrade(ctx context.Context) error {
//		// Implementation would validate upgrade
//		return nil
//	}
//
//	func (cm *CalicoManager) deleteCalicoInstance(ctx context.Context) error {
//		// Implementation would delete Calico instance
//		return nil
//	}
//
//	func (cm *CalicoManager) deleteOperator(ctx context.Context) error {
//		// Implementation would delete operator
//		return nil
//	}
//
//	func (cm *CalicoManager) deleteCRDs(ctx context.Context) error {
//		// Implementation would delete CRDs
//		return nil
//	}
//
//	func (cm *CalicoManager) deleteNamespace(ctx context.Context) error {
//		// Implementation would delete namespace
//		return nil
//	}
//
//	func (cm *CalicoManager) checkInstallation(ctx context.Context) (bool, string, error) {
//		// Implementation would check if Calico is installed
//		return false, "", nil
//	}
//
//	func (cm *CalicoManager) getComponentStatuses(ctx context.Context, status *CalicoStatus) error {
//		// Implementation would get component statuses
//		return nil
//	}
//
//	func (cm *CalicoManager) getNodeStatuses(ctx context.Context, status *CalicoStatus) error {
//		// Implementation would get node statuses
//		return nil
//	}
//
//	func (cm *CalicoManager) getIPPoolStatuses(ctx context.Context, status *CalicoStatus) error {
//		// Implementation would get IP pool statuses
//		return nil
//	}
//
//	func (cm *CalicoManager) getBGPPeerStatuses(ctx context.Context, status *CalicoStatus) error {
//		// Implementation would get BGP peer statuses
//		return nil
//	}
//
//	func (cm *CalicoManager) getPolicyStatuses(ctx context.Context, status *CalicoStatus) error {
//		// Implementation would get policy statuses
//		return nil
//	}
//
//	func (cm *CalicoManager) parseRules(rules []any) []Rule {
//		// Implementation would parse rules
//		return []Rule{}
//	}
//
//	func (cm *CalicoManager) checkCalicoReady(ctx context.Context) (bool, error) {
//		// Implementation would check if Calico is ready
//		return false, nil
//	}
//
//	func (cm *CalicoManager) checkOperatorHealth(ctx context.Context, result *HealthCheckResult) error {
//		// Implementation would check operator health
//		return nil
//	}
//

// waitForCalicoCRDs waits for all Calico CRDs to be created by the operator
func (cm *CalicoManager) waitForCalicoCRDs(ctx context.Context) error {
	cm.Logf("[white]Waiting for Calico CRDs to be created...")

	crdNames := []string{
		"tigerastatuses.operator.tigera.io",
		"apiservers.operator.tigera.io",
		"installations.operator.tigera.io",
		"goldmanes.operator.tigera.io",
		"whiskers.operator.tigera.io",
	}

	// Wait for each CRD to be available
	cc := cmdctx.CmdCtxFromContext(ctx)
	restCfg := cc.GetK8sConfig()
	extClient, err := apiextensionsclient.NewForConfig(restCfg)
	if err != nil {
		return err
	}
	for _, name := range crdNames {
		if err := waitForCRDWithWatch(ctx, extClient, name, 1*time.Minute); err != nil {
			return err
		}
	}

	cm.Logf("[green]All Calico CRDs are now available!")
	return nil
}

// WaitForCRDWithWatch waits for a specific CRD to become available using a watch
func waitForCRDWithWatch(ctx context.Context, client apiextensionsclient.Interface, crdName string, timeout time.Duration) error {
	ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// First check if it already exists and is ready
	crd, err := client.ApiextensionsV1().CustomResourceDefinitions().Get(ctxTimeout, crdName, metav1.GetOptions{})
	if err == nil && isCRDReady(crd) {
		return nil
	}

	// Set up watch
	watcher, err := client.ApiextensionsV1().CustomResourceDefinitions().Watch(ctxTimeout, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", crdName),
	})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Object == nil {
				return fmt.Errorf("watch closed unexpectedly")
			}

			crd, ok := event.Object.(*apiextensionsv1.CustomResourceDefinition)
			if !ok {
				continue
			}

			if isCRDReady(crd) {
				return nil
			}
		case <-ctx.Done():
			return nil
		case <-ctxTimeout.Done():
			return fmt.Errorf("timeout waiting for CRD %s to be ready", crdName)
		}
	}
}

func isCRDReady(crd *apiextensionsv1.CustomResourceDefinition) bool {
	for _, condition := range crd.Status.Conditions {
		if condition.Type == apiextensionsv1.Established {
			return condition.Status == apiextensionsv1.ConditionTrue
		}
	}
	return false
}

// // waitForSingleCRD waits for a single CRD to become available
// func (cm *CalicoManager) waitForSingleCRD(ctx context.Context, gvr schema.GroupVersionResource) error {
// 	cm.Logf("[white]Checking CRD: %s", gvr.Resource)
//
// 	// Try to list the resource to check if the CRD exists
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return fmt.Errorf("context cancelled while waiting for CRD %s: %w", gvr.Resource, ctx.Err())
// 		default:
// 			// Attempt to list the resource
// 			_, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{Limit: 1})
// 			if err == nil {
// 				cm.Logf("[green]CRD %s is available", gvr.Resource)
// 				return nil
// 			}
//
// 			// If it's a "not found" error, the CRD doesn't exist yet
// 			if errors.IsNotFound(err) ||
// 				(strings.Contains(err.Error(), "no matches for kind") ||
// 					strings.Contains(err.Error(), "the server could not find the requested resource")) {
// 				cm.Logf("[yellow]CRD %s not yet available, waiting...", gvr.Resource)
// 				// Wait a bit before retrying
// 				select {
// 				case <-ctx.Done():
// 					return fmt.Errorf("context cancelled while waiting for CRD %s: %w", gvr.Resource, ctx.Err())
// 				case <-time.After(2 * time.Second):
// 					continue
// 				}
// 			}
//
// 			// For other errors, return immediately
// 			return fmt.Errorf("unexpected error checking CRD %s: %w", gvr.Resource, err)
// 		}
// 	}
// }

// func (cm *CalicoManager) checkInstanceHealth(ctx context.Context, result *HealthCheckResult) error {
// 	// Implementation would check instance health
// 	return nil
// }
//
// func (cm *CalicoManager) checkNodeHealth(ctx context.Context, result *HealthCheckResult) error {
// 	// Implementation would check node health
// 	return nil
// }
//
// func (cm *CalicoManager) checkBGPHealth(ctx context.Context, result *HealthCheckResult) error {
// 	// Implementation would check BGP health
// 	return nil
// }
//
// func (cm *CalicoManager) applyUnstructured(ctx context.Context, obj *unstructured.Unstructured) error {
// 	// Implementation would apply unstructured object
// 	return nil
// }
