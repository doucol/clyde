package util

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/doucol/clyde/internal/cmdContext"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type CatcherFunc func(data string)

type DataCatcher struct {
	ctx            context.Context
	catcher        CatcherFunc
	namespace      string
	containerName  string
	portEnvVarName string
	urlPath        string
}

func NewDataCatcher(ctx context.Context, namespace, containerName, portEnvVarName, urlPath string, catcher CatcherFunc) *DataCatcher {
	return &DataCatcher{
		ctx:            ctx,
		namespace:      namespace,
		containerName:  containerName,
		portEnvVarName: portEnvVarName,
		urlPath:        urlPath,
		catcher:        catcher,
	}
}

func (dc *DataCatcher) CatchDataFromSSEStream() error {
	config := cmdContext.K8sConfigFromContext(dc.ctx)

	// URL for the portforward endpoint on the pod
	podName, port, err := GetPodAndEnvVarWithContainerName(dc.ctx, dc.namespace, dc.containerName, dc.portEnvVarName)
	if err != nil {
		return err
	}

	apiURL, _ := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", config.Host, dc.namespace, podName))

	// Dialer for establishing the connection
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiURL)

	// Port mappings (local port: pod port)
	freePort, err := GetFreePort()
	if err != nil {
		return err
	}
	ports := []string{fmt.Sprintf("%d:%s", freePort, port)}

	// Channels for signaling
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	shutdown := false
	go func() {
		// Shutdown the port forwarding when the context is done
		// This will force the ConsumeSSEStream to exit below
		<-dc.ctx.Done()
		shutdown = true
		stopChan <- struct{}{}
	}()

	defer func() {
		// Shutdown the port forwarding in case we've exited for some other reason
		select {
		case stopChan <- struct{}{}:
		default:
		}
	}()

	pf, err := portforward.New(dialer, ports, stopChan, readyChan, io.Discard, io.Discard)
	if err != nil {
		return err
	}

	// Start the port forwarding
	go func() {
		if err := pf.ForwardPorts(); err != nil && !shutdown {
			panic(err)
		}
	}()

	// Wait for the port forwarding to be ready
	<-readyChan

	sseURL := fmt.Sprintf("http://localhost:%d%s", freePort, dc.urlPath)
	if err := ConsumeSSEStream(dc.ctx, sseURL, dc.catcher); err != nil && !shutdown {
		return err
	}

	return nil
}
