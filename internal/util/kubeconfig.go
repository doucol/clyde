package util

import (
	"fmt"
	"sort"

	"k8s.io/client-go/tools/clientcmd"
)

const (
	KubeconfigSourceFlag    = "--kubeconfig flag"
	KubeconfigSourceEnv     = "KUBECONFIG env var"
	KubeconfigSourceDefault = "default (~/.kube/config)"
)

type KubeconfigInfo struct {
	Path           string
	Source         string
	Contexts       []string
	CurrentContext string
}

func LoadKubeconfigInfo(path, source string) (*KubeconfigInfo, error) {
	if path == "" {
		return nil, fmt.Errorf("kubeconfig path is empty")
	}
	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig %q: %w", path, err)
	}
	names := make([]string, 0, len(cfg.Contexts))
	for name := range cfg.Contexts {
		names = append(names, name)
	}
	sort.Strings(names)
	return &KubeconfigInfo{
		Path:           path,
		Source:         source,
		Contexts:       names,
		CurrentContext: cfg.CurrentContext,
	}, nil
}
