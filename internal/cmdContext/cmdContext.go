package cmdContext

import (
	"context"

	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

type cmdContextKeyType string

const cmdContextKey cmdContextKeyType = "CmdContextKey"

type CmdContext struct {
	KubeConfig  string
	KubeContext string
	dc          *dynamic.DynamicClient
}

func (c *CmdContext) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, cmdContextKey, c)
}

func (c *CmdContext) Client() *dynamic.DynamicClient {
	if c.dc != nil {
		return c.dc
	}

	var configOverrides *clientcmd.ConfigOverrides
	configLoadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: c.KubeConfig}
	if c.KubeContext != "" {
		configOverrides = &clientcmd.ConfigOverrides{CurrentContext: c.KubeContext}
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

func CmdContextFromContext(ctx context.Context) *CmdContext {
	return ctx.Value(cmdContextKey).(*CmdContext)
}

func ClientFromContext(ctx context.Context) *dynamic.DynamicClient {
	return CmdContextFromContext(ctx).Client()
}

func NewCmdContext(kubeConfig, kubeContext string) *CmdContext {
	return &CmdContext{KubeConfig: kubeConfig, KubeContext: kubeContext, dc: nil}
}
