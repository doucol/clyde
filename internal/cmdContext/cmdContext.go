package cmdContext

import (
	"context"
	"errors"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type cmdContextKeyType string

const cmdContextKey cmdContextKeyType = "CmdContextKey"

type CmdContext struct {
	kubeConfig  string
	kubeContext string
	k8scfg      *rest.Config
	dc          *dynamic.DynamicClient
	cs          *kubernetes.Clientset
	cancel      context.CancelFunc
}

func (c *CmdContext) ToContext(ctx context.Context) context.Context {
	ctx2, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	return context.WithValue(ctx2, cmdContextKey, c)
}

func (c *CmdContext) GetConfig() *rest.Config {
	if c.k8scfg != nil {
		return c.k8scfg
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
	c.k8scfg = config
	return c.k8scfg
}

func (c *CmdContext) ClientDyn() *dynamic.DynamicClient {
	if c.dc != nil {
		return c.dc
	}
	config := c.GetConfig()
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
	config := c.GetConfig()
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	c.cs = cs
	return c.cs
}

func (c *CmdContext) Cancel() {
	if c.cancel == nil {
		panic(errors.New("cancel function not set"))
	}
	c.cancel()
}

func CmdContextFromContext(ctx context.Context) *CmdContext {
	return ctx.Value(cmdContextKey).(*CmdContext)
}

func ClientDynFromContext(ctx context.Context) *dynamic.DynamicClient {
	return CmdContextFromContext(ctx).ClientDyn()
}

func ClientsetFromContext(ctx context.Context) *kubernetes.Clientset {
	return CmdContextFromContext(ctx).Clientset()
}

func K8sConfigFromContext(ctx context.Context) *rest.Config {
	return CmdContextFromContext(ctx).GetConfig()
}

func NewCmdContext(kubeConfig, kubeContext string) *CmdContext {
	return &CmdContext{kubeConfig: kubeConfig, kubeContext: kubeContext}
}
