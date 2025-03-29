package cmdctx

import (
	"context"
	"errors"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type cmdCtxKeyType string

const cmdCtxKey cmdCtxKeyType = "CmdContextKey"

type CmdCtx struct {
	kubeConfig  string
	kubeContext string
	k8scfg      *rest.Config
	dc          *dynamic.DynamicClient
	cs          *kubernetes.Clientset
	cancel      context.CancelFunc
}

func NewCmdCtx(kubeConfig, kubeContext string) *CmdCtx {
	return &CmdCtx{kubeConfig: kubeConfig, kubeContext: kubeContext}
}

func (c *CmdCtx) ToContext(ctx context.Context) context.Context {
	newctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	return context.WithValue(newctx, cmdCtxKey, c)
}

func (c *CmdCtx) GetK8sConfig() *rest.Config {
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

func (c *CmdCtx) ClientDyn() *dynamic.DynamicClient {
	if c.dc != nil {
		return c.dc
	}
	config := c.GetK8sConfig()
	dc, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	c.dc = dc
	return c.dc
}

func (c *CmdCtx) Clientset() *kubernetes.Clientset {
	if c.cs != nil {
		return c.cs
	}
	config := c.GetK8sConfig()
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	c.cs = cs
	return c.cs
}

func (c *CmdCtx) Cancel() {
	if c.cancel == nil {
		panic(errors.New("cancel function not set"))
	}
	c.cancel()
}

func CmdCtxFromContext(ctx context.Context) *CmdCtx {
	return ctx.Value(cmdCtxKey).(*CmdCtx)
}

func K8sClientDynFromContext(ctx context.Context) *dynamic.DynamicClient {
	return CmdCtxFromContext(ctx).ClientDyn()
}

func K8sClientsetFromContext(ctx context.Context) *kubernetes.Clientset {
	return CmdCtxFromContext(ctx).Clientset()
}

func K8sConfigFromContext(ctx context.Context) *rest.Config {
	return CmdCtxFromContext(ctx).GetK8sConfig()
}
