package cmdContext

import (
	"context"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"
)

type cmdContextKeyType string

const cmdContextKey cmdContextKeyType = "CmdContextKey"

type CmdContext struct {
	KubeConfig  string
	KubeContext string
	clientset   *kubernetes.Clientset
}

func (c *CmdContext) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, cmdContextKey, c)
}

func (c *CmdContext) Clientset() *kubernetes.Clientset {
	if c.clientset != nil {
		return c.clientset
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

	// create the clientset
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	c.clientset = cs
	return c.clientset
}

func CmdContextFromContext(ctx context.Context) *CmdContext {
	return ctx.Value(cmdContextKey).(*CmdContext)
}

func ClientsetFromContext(ctx context.Context) *kubernetes.Clientset {
	return CmdContextFromContext(ctx).Clientset()
}

func NewCmdContext(kubeConfig, kubeContext string) *CmdContext {
	return &CmdContext{KubeConfig: kubeConfig, KubeContext: kubeContext, clientset: nil}
}
