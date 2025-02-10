package cmdContext

import (
	"context"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

type cmdContextKeyType string

const cmdContextKey cmdContextKeyType = "CmdContextKey"

type CmdContext struct {
	kubeConfig  string
	kubeContext string
	dc          *dynamic.DynamicClient
	cs          *kubernetes.Clientset
}

func (c *CmdContext) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, cmdContextKey, c)
}

func (c *CmdContext) ClientDyn() *dynamic.DynamicClient {
	if c.dc != nil {
		return c.dc
	}

	var configOverrides *clientcmd.ConfigOverrides
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: c.kubeConfig}
	if c.kubeContext != "" {
		configOverrides = &clientcmd.ConfigOverrides{CurrentContext: c.kubeContext}
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, configOverrides).ClientConfig()
	if err != nil {
		panic(err)
	}

	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	c.dc = dc
	return c.dc
}

func (c *CmdContext) Clientset() *kubernetes.Clientset {
	if c.cs != nil {
		return c.cs
	}

	var configOverrides *clientcmd.ConfigOverrides
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: c.kubeConfig}
	if c.kubeContext != "" {
		configOverrides = &clientcmd.ConfigOverrides{CurrentContext: c.kubeContext}
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, configOverrides).ClientConfig()
	if err != nil {
		panic(err)
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	c.cs = cs
	return c.cs
}

func CmdContextFromContext(ctx context.Context) *CmdContext {
	return ctx.Value(cmdContextKey).(*CmdContext)
}

func ClientFromContext(ctx context.Context) *dynamic.DynamicClient {
	return CmdContextFromContext(ctx).ClientDyn()
}

func NewCmdContext(kubeConfig, kubeContext string) *CmdContext {
	return &CmdContext{kubeConfig: kubeConfig, kubeContext: kubeContext, dc: nil}
}
