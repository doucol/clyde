// Package calico provides comprehensive Calico OSS management capabilities
// including installation, configuration, and management of Calico resources.
package calico

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/doucol/clyde/internal/cmdctx"
	"github.com/doucol/clyde/internal/githubversions"
	k8sapplier "github.com/doucol/clyde/internal/k8sapply"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
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
func (cm *CalicoManager) Install(ctx context.Context) error {
	latestVersion, err := githubversions.GetLatestStableSemverTag(ctx, "projectcalico", "calico")
	if err != nil {
		cm.Logf("[red]Failed to get latest Calico version: %v", err)
		return err
	}

	cm.Logf("[white]Installing latest Calico version: %s", latestVersion)

	if err := cm.installOperator(ctx, latestVersion); err != nil {
		cm.Logf("[red]Failed to install Calico operator: %s", err.Error())
		return fmt.Errorf("failed to install Calico operator: %w", err)
	}

	cm.Logf("[green]Calico installation completed successfully!")
	return nil
}

func (cm *CalicoManager) applyAndLog(ctx context.Context, url string) error {
	applier, err := k8sapplier.NewApplierWithClients(cm.clientset, cm.dynamic, nil)
	if err != nil {
		return fmt.Errorf("failed to create k8s applier: %w", err)
	}
	defer applier.Close()

	results, err := applier.ApplyFromURL(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("failed to apply kubernetes resources: %w", err)
	}
	error := false
	for _, result := range results {
		ns := result.Namespace
		if ns != "" {
			ns = ns + "/"
		}
		if result.Error != nil {
			error = true
			cm.Logf("[red]Error applying %s%s (%s): %v", ns, result.Name, result.Kind, result.Error)
		} else {
			cm.Logf("[white]Successfully %s: %s%s (%s)", result.Action, ns, result.Name, result.Kind)
		}
	}
	if error {
		return fmt.Errorf("some manifests failed to apply")
	}
	return nil
}

func (cm *CalicoManager) GoldmaneWhiskerAvailable() (bool, error) {
	// Use the discovery client to determine if the tigera secure specific APIs exist.
	resources, err := cm.clientset.Discovery().ServerResourcesForGroupVersion("operator.tigera.io/v1")
	if err != nil {
		return false, err
	}
	goldmaneFound := false
	whiskerFound := false
	for _, r := range resources.APIResources {
		switch r.Kind {
		case "Goldmane":
			goldmaneFound = true
		case "Whisker":
			whiskerFound = true
		}
	}
	return (goldmaneFound && whiskerFound), nil
}

// GoldmaneWhiskerAccessible checks if Goldmane and Whisker resources are actually accessible
func (cm *CalicoManager) GoldmaneWhiskerAccessible(ctx context.Context) (bool, error) {
	// Check if we can actually access the resources, not just discover them
	goldmaneAccessible := false
	whiskerAccessible := false

	// Check Goldmane resource accessibility
	goldmaneGVR := schema.GroupVersionResource{
		Group:    "operator.tigera.io",
		Version:  "v1",
		Resource: "goldmanes",
	}

	_, err := cm.dynamic.Resource(goldmaneGVR).List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil {
		goldmaneAccessible = true
	} else if strings.Contains(err.Error(), "no matches for kind") {
		// Resource not ready yet
		goldmaneAccessible = false
	} else {
		// Other error, log it but don't fail
		cm.Logf("[yellow]Warning: Error checking Goldmane accessibility: %v", err)
		goldmaneAccessible = false
	}

	// Check Whisker resource accessibility
	whiskerGVR := schema.GroupVersionResource{
		Group:    "operator.tigera.io",
		Version:  "v1",
		Resource: "whiskers",
	}

	_, err = cm.dynamic.Resource(whiskerGVR).List(ctx, metav1.ListOptions{Limit: 1})
	if err == nil {
		whiskerAccessible = true
	} else if strings.Contains(err.Error(), "no matches for kind") {
		// Resource not ready yet
		whiskerAccessible = false
	} else {
		// Other error, log it but don't fail
		cm.Logf("[yellow]Warning: Error checking Whisker accessibility: %v", err)
		whiskerAccessible = false
	}

	return (goldmaneAccessible && whiskerAccessible), nil
}

// WaitForGoldmaneWhiskerAvailable waits for Goldmane and Whisker resources to be available
func (cm *CalicoManager) WaitForGoldmaneWhiskerAvailable(ctx context.Context, timeout time.Duration) error {
	cm.Logf("[white]Waiting for Goldmane and Whisker resources to be available...")

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll until both resources are available
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("timeout waiting for Goldmane and Whisker resources: %w", ctxWithTimeout.Err())
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			available, err := cm.GoldmaneWhiskerAvailable()
			if err != nil {
				cm.Logf("[yellow]Error checking Goldmane and Whisker availability: %v", err)
				continue
			}

			if available {
				cm.Logf("[green]Goldmane and Whisker resources are available!")
				return nil
			}

			cm.Logf("[white]Goldmane and Whisker resources not available yet, waiting...")
		}
	}
}

// WaitForGoldmaneWhiskerAccessible waits for Goldmane and Whisker resources to be accessible
func (cm *CalicoManager) WaitForGoldmaneWhiskerAccessible(ctx context.Context, timeout time.Duration) error {
	cm.Logf("[white]Waiting for Goldmane and Whisker resources to be accessible...")

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll until both resources are accessible
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("timeout waiting for Goldmane and Whisker resources to be accessible: %w", ctxWithTimeout.Err())
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			accessible, err := cm.GoldmaneWhiskerAccessible(ctxWithTimeout)
			if err != nil {
				cm.Logf("[yellow]Error checking Goldmane and Whisker accessibility: %v", err)
				continue
			}

			if accessible {
				cm.Logf("[green]Goldmane and Whisker resources are accessible!")
				return nil
			}

			cm.Logf("[white]Goldmane and Whisker resources not accessible yet, waiting...")
		}
	}
}

// WaitForGoldmaneWhiskerAvailableWithRetry waits for Goldmane and Whisker with retry mechanism
func (cm *CalicoManager) WaitForGoldmaneWhiskerAvailableWithRetry(ctx context.Context, maxRetries int, initialBackoff time.Duration) error {
	backoff := initialBackoff

	for i := 0; i < maxRetries; i++ {
		cm.Logf("[white]Attempt %d/%d: Waiting for Goldmane and Whisker resources...", i+1, maxRetries)

		err := cm.WaitForGoldmaneWhiskerAvailable(ctx, 2*time.Minute)
		if err == nil {
			cm.Logf("[green]Goldmane and Whisker resources are available on attempt %d", i+1)
			return nil
		}

		if i < maxRetries-1 {
			cm.Logf("[yellow]Attempt %d failed: %v. Retrying in %v...", i+1, err, backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		} else {
			cm.Logf("[red]All %d attempts failed. Goldmane and Whisker resources are not available: %v", maxRetries, err)
			return fmt.Errorf("exceeded maximum retries waiting for Goldmane and Whisker resources: %w", err)
		}
	}

	return fmt.Errorf("exceeded maximum retries waiting for Goldmane and Whisker resources")
}

// WaitForGoldmaneWhiskerCompletelyReady waits for Goldmane and Whisker resources to be both available and accessible
func (cm *CalicoManager) WaitForGoldmaneWhiskerCompletelyReady(ctx context.Context, timeout time.Duration) error {
	cm.Logf("[white]Waiting for Goldmane and Whisker resources to be completely ready...")

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// First wait for the resources to be available (discoverable)
	cm.Logf("[white]Step 1: Waiting for Goldmane and Whisker resources to be discoverable...")
	if err := cm.WaitForGoldmaneWhiskerAvailable(ctxWithTimeout, timeout/2); err != nil {
		return fmt.Errorf("goldmane and Whisker resources not discoverable: %w", err)
	}

	// Then wait for them to be accessible (usable)
	cm.Logf("[white]Step 2: Waiting for Goldmane and Whisker resources to be accessible...")
	if err := cm.WaitForGoldmaneWhiskerAccessible(ctxWithTimeout, timeout/2); err != nil {
		return fmt.Errorf("goldmane and Whisker resources not accessible: %w", err)
	}

	cm.Logf("[green]Goldmane and Whisker resources are completely ready!")
	return nil
}

// WaitForGoldmaneWhiskerCompletelyReadyWithRetry waits for Goldmane and Whisker with retry mechanism
func (cm *CalicoManager) WaitForGoldmaneWhiskerCompletelyReadyWithRetry(ctx context.Context, maxRetries int, initialBackoff time.Duration) error {
	backoff := initialBackoff

	for i := 0; i < maxRetries; i++ {
		cm.Logf("[white]Attempt %d/%d: Waiting for Goldmane and Whisker resources to be completely ready...", i+1, maxRetries)

		err := cm.WaitForGoldmaneWhiskerCompletelyReady(ctx, 3*time.Minute)
		if err == nil {
			cm.Logf("[green]Goldmane and Whisker resources are completely ready on attempt %d", i+1)
			return nil
		}

		if i < maxRetries-1 {
			cm.Logf("[yellow]Attempt %d failed: %v. Retrying in %v...", i+1, err, backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		} else {
			cm.Logf("[red]All %d attempts failed. Goldmane and Whisker resources are not completely ready: %v", maxRetries, err)
			return fmt.Errorf("exceeded maximum retries waiting for Goldmane and Whisker resources: %w", err)
		}
	}

	return fmt.Errorf("exceeded maximum retries waiting for Goldmane and Whisker resources")
}

// installOperator installs the Calico operator and custom resources
func (cm *CalicoManager) installOperator(ctx context.Context, latestVersion string) error {
	cc := cmdctx.CmdCtxFromContext(ctx)
	cm.Logf("[white]Installing Calico operator...")

	cm.Logf("[white]Applying Calico operator manifests from version %s", latestVersion)
	murl := fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%s/manifests/tigera-operator.yaml", latestVersion)
	if err := cm.applyAndLog(ctx, murl); err != nil {
		cm.Logf("[red]Failed to install Calico operator: %v", err)
		return err
	}

	// Wait for the Calico CRDs to be created by the operator with retry mechanism
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		cm.Logf("[white]Attempt %d/%d: Waiting for Calico CRDs to be ready...", attempt, maxRetries)

		if err := cm.waitForCalicoCRDs(ctx); err != nil {
			if attempt == maxRetries {
				cm.Logf("[red]All attempts failed. Calico CRDs did not become ready: %s", err.Error())
				return fmt.Errorf("failed to wait for Calico CRDs after %d attempts: %w", maxRetries, err)
			}

			cm.Logf("[yellow]Attempt %d failed: %v. Retrying in 10 seconds...", attempt, err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Second):
				continue
			}
		}

		cm.Logf("[green]Calico CRDs are ready on attempt %d", attempt)
		break
	}

	// Wait for Tigera operator to be ready
	cm.Log("[white]Waiting for all Tigera operator statuses to be ready")
	err := cm.WaitForTigeraOperatorAvailable(ctx, cc.Clientset(), cc.ClientDyn(), 5*time.Minute)
	if err != nil {
		cm.Logf("[red]Tigera operator did not become ready: %s", err.Error())
		return err
	}

	cm.Logf("[white]Applying Calico custom resources from version %s", latestVersion)
	murl = fmt.Sprintf("https://raw.githubusercontent.com/projectcalico/calico/%s/manifests/custom-resources.yaml", latestVersion)
	maxRetries = 100
	success := false
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := cm.applyAndLog(ctx, murl); err != nil {
			time.Sleep(5 * time.Second) // Wait before retrying
		} else {
			success = true
			cm.Log("[green]Successfully applied custom resources for Calico operator")
			break
		}
	}
	if !success {
		return fmt.Errorf("failed to apply custom resources for Calico operator after %d attempts", maxRetries)
	}
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

// waitForCalicoCRDs waits for all Calico CRDs to be created by the operator
func (cm *CalicoManager) waitForCalicoCRDs(ctx context.Context) error {
	cm.Logf("[white]Waiting for Calico CRDs to be created...")

	cc := cmdctx.CmdCtxFromContext(ctx)
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(cc.GetK8sConfig())
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %v", err)
	}

	// Define the CRDs we need to wait for, with specific focus on Goldmane and Whisker
	crdNames := []string{
		"installations.operator.tigera.io",
		"goldmanes.operator.tigera.io",
		"whiskers.operator.tigera.io",
	}

	// Wait for all CRDs to be established first
	cm.Logf("[white]Waiting for all Calico CRDs to be established...")
	for _, name := range crdNames {
		cm.Logf("[white]Waiting for CRD %s to be established...", name)
		if err := WaitForCRDEstablished(ctx, apiextensionsClient, name, 5*time.Minute); err != nil {
			cm.Logf("[red]CRD %s did not become established: %v", name, err)
			return fmt.Errorf("CRD %s did not become established: %w", name, err)
		}
		cm.Logf("[green]CRD %s is established", name)
	}

	// Additional wait for Goldmane and Whisker CRDs to be completely ready
	cm.Logf("[white]Waiting for Goldmane and Whisker CRDs to be completely ready...")

	// Use the comprehensive waiting function that checks both availability and accessibility
	if err := cm.WaitForGoldmaneWhiskerCompletelyReady(ctx, 5*time.Minute); err != nil {
		cm.Logf("[red]Goldmane and Whisker CRDs did not become completely ready: %v", err)
		return fmt.Errorf("Goldmane and Whisker CRDs did not become completely ready: %w", err)
	}

	// Wait for the API server to be ready to serve the new CRDs
	if err := cm.waitForAPIServerReady(ctx); err != nil {
		cm.Logf("[red]API server is not ready to serve new CRDs: %v", err)
		return fmt.Errorf("API server is not ready to serve new CRDs: %w", err)
	}

	// Critical: Wait for REST mappings to be available
	// This is what causes the "no matches for kind" error
	cm.Logf("[white]Waiting for REST mappings to be available...")
	if err := cm.waitForRESTMappingsAvailable(ctx, crdNames); err != nil {
		cm.Logf("[red]REST mappings are not available: %v", err)
		return fmt.Errorf("REST mappings are not available: %w", err)
	}

	// Final verification that all CRDs are accessible
	cm.Logf("[white]Verifying all CRDs are accessible...")
	if err := cm.verifyAllCRDsAccessible(ctx, crdNames); err != nil {
		cm.Logf("[red]CRD accessibility verification failed: %v", err)
		return fmt.Errorf("CRD accessibility verification failed: %w", err)
	}

	cm.Logf("[green]All Calico CRDs are now available, ready, and accessible!")
	return nil
}

// waitForRESTMappingsAvailable waits for the API server to have REST mappings available for the CRDs
// This is critical to prevent "no matches for kind" errors
func (cm *CalicoManager) waitForRESTMappingsAvailable(ctx context.Context, crdNames []string) error {
	cm.Logf("[white]Waiting for REST mappings to be available for all CRDs...")

	// Wait for each CRD to have REST mappings available
	for _, crdName := range crdNames {
		cm.Logf("[white]Waiting for REST mapping for CRD %s...", crdName)

		// Use a longer timeout for REST mapping availability
		timeout := 2 * time.Minute
		if err := cm.waitForSingleCRDRESTMapping(ctx, crdName, timeout); err != nil {
			return fmt.Errorf("REST mapping for CRD %s is not available: %w", crdName, err)
		}

		cm.Logf("[green]REST mapping for CRD %s is available", crdName)
	}

	return nil
}

// waitForCRDReady waits for a specific CRD to be completely ready and established
func (cm *CalicoManager) waitForCRDReady(ctx context.Context, crdName string, timeout time.Duration) error {
	cc := cmdctx.CmdCtxFromContext(ctx)
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(cc.GetK8sConfig())
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %v", err)
	}

	// First wait for the CRD to be established
	if err := WaitForCRDEstablished(ctx, apiextensionsClient, crdName, timeout); err != nil {
		return fmt.Errorf("CRD %s did not become established: %w", crdName, err)
	}

	// Additional wait to ensure the CRD is completely ready
	// This includes waiting for the API server to fully recognize the CRD
	cm.Logf("[white]Ensuring CRD %s is completely ready...", crdName)

	// Wait a bit more to ensure the API server has fully processed the CRD
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
		// Continue after the additional wait
	}

	// Verify the CRD is accessible by trying to list resources
	// This ensures the API server can actually serve requests for this CRD
	if err := cm.verifyCRDAccessibility(ctx, crdName); err != nil {
		return fmt.Errorf("CRD %s is not yet accessible: %w", crdName, err)
	}

	return nil
}

// waitForSingleCRDRESTMapping waits for a single CRD to have REST mappings available
func (cm *CalicoManager) waitForSingleCRDRESTMapping(ctx context.Context, crdName string, timeout time.Duration) error {
	// Parse the CRD name to get group, version, and resource
	parts := strings.Split(crdName, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid CRD name format: %s", crdName)
	}

	resource := parts[0]
	group := strings.Join(parts[1:], ".")

	// Try different API versions that Calico CRDs commonly use
	versions := []string{"v1", "v1alpha1", "v1beta1"}

	// Create a context with timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Poll until the REST mapping is available
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctxWithTimeout.Done():
			return fmt.Errorf("timeout waiting for REST mapping for CRD %s", crdName)
		case <-ticker.C:
			// Try to access the resource to check if REST mapping is available
			for _, version := range versions {
				gvr := schema.GroupVersionResource{
					Group:    group,
					Version:  version,
					Resource: resource,
				}

				// Try to list resources - this will fail with "no matches for kind" if REST mapping isn't ready
				_, err := cm.dynamic.Resource(gvr).List(ctxWithTimeout, metav1.ListOptions{Limit: 1})
				if err == nil {
					// Success! REST mapping is available
					return nil
				}

				// Check if it's a "no matches for kind" error
				if strings.Contains(err.Error(), "no matches for kind") {
					// REST mapping not ready yet, continue waiting
					continue
				}

				// For other errors, the REST mapping might be ready but there's another issue
				// Log it but don't fail immediately
				cm.Logf("[yellow]Warning: Error checking REST mapping for CRD %s with version %s: %v", crdName, version, err)
			}
		}
	}
}

// verifyCRDAccessibility verifies that a CRD is accessible by the API server
func (cm *CalicoManager) verifyCRDAccessibility(ctx context.Context, crdName string) error {
	// Parse the CRD name to get group, version, and resource
	// Format: resource.group
	parts := strings.Split(crdName, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid CRD name format: %s", crdName)
	}

	resource := parts[0]
	group := strings.Join(parts[1:], ".")

	// Try different API versions that Calico CRDs commonly use
	versions := []string{"v1", "v1alpha1", "v1beta1"}

	for _, version := range versions {
		gvr := schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: resource,
		}

		// Attempt to list resources with a limit of 1 to verify accessibility
		_, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{Limit: 1})
		if err == nil {
			// Success! The CRD is accessible with this version
			cm.Logf("[white]CRD %s is accessible via %s API", crdName, version)
			return nil
		}

		// Check for specific REST mapping errors
		if strings.Contains(err.Error(), "no matches for kind") {
			// This is the specific error we're trying to prevent
			cm.Logf("[yellow]REST mapping not ready for CRD %s with version %s: %v", crdName, version, err)
			continue
		}

		// If we get a "not found" error, the CRD might not be ready yet
		if strings.Contains(err.Error(), "the server could not find the requested resource") ||
			strings.Contains(err.Error(), "not found") {
			// This version doesn't exist, try the next one
			continue
		}

		// For other errors, log them but continue trying other versions
		cm.Logf("[yellow]Warning: Error checking CRD %s with version %s: %v", crdName, version, err)
		continue
	}

	// If we get here, none of the versions worked
	return fmt.Errorf("CRD %s is not accessible by the API server with any supported version (%s). This may indicate REST mapping issues.", crdName, strings.Join(versions, ", "))
}

// verifyAllCRDsAccessible verifies that all specified CRDs are accessible by the API server.
func (cm *CalicoManager) verifyAllCRDsAccessible(ctx context.Context, crdNames []string) error {
	for _, crdName := range crdNames {
		if err := cm.verifyCRDAccessibility(ctx, crdName); err != nil {
			return fmt.Errorf("CRD %s is not accessible: %w", crdName, err)
		}
	}
	return nil
}

// waitForAPIServerReady waits for the API server to be ready to serve the new CRDs
// This is often necessary after CRDs are created to ensure REST mappings are available
func (cm *CalicoManager) waitForAPIServerReady(ctx context.Context) error {
	cm.Logf("[white]Waiting for API server to be ready to serve new CRDs...")

	// Wait for the API server to process the new CRDs
	// This is a common issue where CRDs are established but not yet available via the API
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(15 * time.Second):
		// Wait 15 seconds for the API server to process new CRDs
		cm.Logf("[white]Waited 15 seconds for API server to process new CRDs")
	}

	cm.Logf("[green]API server should be ready to serve new CRDs")
	return nil
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

// DiagnoseCRDIssues provides detailed diagnostic information about CRD readiness issues
func (cm *CalicoManager) DiagnoseCRDIssues(ctx context.Context) error {
	cm.Logf("[white]Diagnosing CRD readiness issues...")

	cc := cmdctx.CmdCtxFromContext(ctx)
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(cc.GetK8sConfig())
	if err != nil {
		return fmt.Errorf("failed to create apiextensions client: %v", err)
	}

	crdNames := []string{
		"installations.operator.tigera.io",
		"goldmanes.operator.tigera.io",
		"whiskers.operator.tigera.io",
	}

	cm.Logf("[white]Checking status of all Calico CRDs...")

	for _, crdName := range crdNames {
		status, err := GetCRDStatus(ctx, apiextensionsClient, crdName)
		if err != nil {
			cm.Logf("[red]Failed to get status for CRD %s: %v", crdName, err)
			continue
		}

		cm.Logf("[white]CRD: %s", crdName)
		cm.Logf("[white]  Status: %s", status.GetStatusSummary())
		cm.Logf("[white]  Established: %t", status.Established)
		cm.Logf("[white]  Names Accepted: %t", status.NamesAccepted)

		if status.LastTransitionTime != nil {
			cm.Logf("[white]  Last Transition: %s", status.LastTransitionTime.Format(time.RFC3339))
		}

		// Show all conditions
		for condType, condStatus := range status.Conditions {
			cm.Logf("[white]  Condition %s: %s", condType, condStatus)
		}

		cm.Logf("") // Empty line for readability
	}

	// Check REST mapping availability
	cm.Logf("[white]Checking REST mapping availability...")
	for _, crdName := range crdNames {
		cm.Logf("[white]Checking REST mapping for CRD %s...", crdName)

		// Parse CRD name
		parts := strings.Split(crdName, ".")
		if len(parts) < 2 {
			cm.Logf("[red]Invalid CRD name format: %s", crdName)
			continue
		}

		resource := parts[0]
		group := strings.Join(parts[1:], ".")

		// Check different API versions
		versions := []string{"v1", "v1alpha1", "v1beta1"}
		restMappingAvailable := false

		for _, version := range versions {
			gvr := schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: resource,
			}

			// Try to list resources to check REST mapping
			_, err := cm.dynamic.Resource(gvr).List(ctx, metav1.ListOptions{Limit: 1})
			if err == nil {
				cm.Logf("[green]  REST mapping available via %s API", version)
				restMappingAvailable = true
				break
			} else if strings.Contains(err.Error(), "no matches for kind") {
				cm.Logf("[red]  REST mapping NOT available via %s API: %v", version, err)
			} else {
				cm.Logf("[yellow]  Error checking %s API: %v", version, err)
			}
		}

		if restMappingAvailable {
			cm.Logf("[green]  REST mapping: AVAILABLE", crdName)
		} else {
			cm.Logf("[red]  REST mapping: NOT AVAILABLE", crdName)
		}

		cm.Logf("") // Empty line for readability
	}

	// Check if the Tigera operator is running
	cm.Logf("[white]Checking Tigera operator deployment status...")
	deployment, err := cm.clientset.AppsV1().Deployments("tigera-operator").Get(ctx, "tigera-operator", metav1.GetOptions{})
	if err != nil {
		cm.Logf("[red]Failed to get Tigera operator deployment: %v", err)
	} else {
		cm.Logf("[white]Tigera operator deployment:")
		cm.Logf("[white]  Replicas: %d/%d", deployment.Status.ReadyReplicas, *deployment.Spec.Replicas)
		cm.Logf("[white]  Available: %d", deployment.Status.AvailableReplicas)
		cm.Logf("[white]  Updated: %d", deployment.Status.UpdatedReplicas)

		// Check pod status
		pods, err := cm.clientset.CoreV1().Pods("tigera-operator").List(ctx, metav1.ListOptions{
			LabelSelector: "name=tigera-operator",
		})
		if err != nil {
			cm.Logf("[red]Failed to get Tigera operator pods: %v", err)
		} else {
			cm.Logf("[white]Tigera operator pods:")
			for _, pod := range pods.Items {
				phase := pod.Status.Phase
				ready := "Not Ready"
				for _, condition := range pod.Status.Conditions {
					if condition.Type == "Ready" {
						if condition.Status == "True" {
							ready = "Ready"
						}
						break
					}
				}
				cm.Logf("[white]  %s: %s (%s)", pod.Name, phase, ready)
			}
		}
	}

	return nil
}

// DiagnoseGoldmaneWhiskerIssues provides detailed diagnostic information about Goldmane and Whisker resource issues
func (cm *CalicoManager) DiagnoseGoldmaneWhiskerIssues(ctx context.Context) error {
	cm.Logf("[white]Diagnosing Goldmane and Whisker resource issues...")

	// Check discovery availability
	cm.Logf("[white]Checking discovery availability...")
	available, err := cm.GoldmaneWhiskerAvailable()
	if err != nil {
		cm.Logf("[red]Failed to check discovery availability: %v", err)
	} else {
		cm.Logf("[white]Discovery availability: %t", available)
	}

	// Check resource accessibility
	cm.Logf("[white]Checking resource accessibility...")
	accessible, err := cm.GoldmaneWhiskerAccessible(ctx)
	if err != nil {
		cm.Logf("[red]Failed to check resource accessibility: %v", err)
	} else {
		cm.Logf("[white]Resource accessibility: %t", accessible)
	}

	// Check CRD status
	cm.Logf("[white]Checking CRD status...")
	cc := cmdctx.CmdCtxFromContext(ctx)
	apiextensionsClient, err := apiextensionsclientset.NewForConfig(cc.GetK8sConfig())
	if err != nil {
		cm.Logf("[red]Failed to create apiextensions client: %v", err)
	} else {
		crdNames := []string{"goldmanes.operator.tigera.io", "whiskers.operator.tigera.io"}

		for _, crdName := range crdNames {
			status, err := GetCRDStatus(ctx, apiextensionsClient, crdName)
			if err != nil {
				cm.Logf("[red]Failed to get status for CRD %s: %v", crdName, err)
				continue
			}

			cm.Logf("[white]CRD: %s", crdName)
			cm.Logf("[white]  Status: %s", status.GetStatusSummary())
			cm.Logf("[white]  Established: %t", status.Established)
			cm.Logf("[white]  Names Accepted: %t", status.NamesAccepted)

			if status.LastTransitionTime != nil {
				cm.Logf("[white]  Last Transition: %s", status.LastTransitionTime.Format(time.RFC3339))
			}

			// Show all conditions
			for condType, condStatus := range status.Conditions {
				cm.Logf("[white]  Condition %s: %s", condType, condStatus)
			}

			cm.Logf("") // Empty line for readability
		}
	}

	// Check if the resources can be listed
	cm.Logf("[white]Checking resource listing capability...")

	// Try to list Goldmane resources
	goldmaneGVR := schema.GroupVersionResource{
		Group:    "operator.tigera.io",
		Version:  "v1",
		Resource: "goldmanes",
	}

	list, err := cm.dynamic.Resource(goldmaneGVR).List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		cm.Logf("[red]Failed to list Goldmane resources: %v", err)
	} else {
		cm.Logf("[green]Successfully listed Goldmane resources: %d found", len(list.Items))
	}

	// Try to list Whisker resources
	whiskerGVR := schema.GroupVersionResource{
		Group:    "operator.tigera.io",
		Version:  "v1",
		Resource: "whiskers",
	}

	list, err = cm.dynamic.Resource(whiskerGVR).List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		cm.Logf("[red]Failed to list Whisker resources: %v", err)
	} else {
		cm.Logf("[green]Successfully listed Whisker resources: %d found", len(list.Items))
	}

	return nil
}
