package whisker

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/util"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	CalicoNamespace             = "calico-system"
	WhiskerBackendContainerName = "whisker-backend"
)

func WatchFlows(ctx context.Context) error {
	config := cmdContext.K8sConfigFromContext(ctx)

	// URL for the portforward endpoint on the pod
	podName, port, err := util.GetPodAndEnvVarWithContainerName(ctx, CalicoNamespace, WhiskerBackendContainerName, "PORT")
	if err != nil {
		return err
	}

	apiURL, _ := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/calico-system/pods/%s/portforward", config.Host, podName))

	// Dialer for establishing the connection
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return err
	}
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiURL)

	// Port mappings (local port: pod port)
	freePort, err := util.GetFreePort()
	if err != nil {
		return err
	}
	ports := []string{fmt.Sprintf("%d:%s", freePort, port)}

	// Channels for signaling
	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})

	// Create the port forwarder
	forwarder, err := portforward.New(dialer, ports, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	defer forwarder.Close()

	// Start the port forwarding
	go func() {
		if err := forwarder.ForwardPorts(); err != nil {
			fmt.Println(err)
		}
	}()

	// Wait for the port forwarding to be ready
	<-readyChan
	// fmt.Println("Port forwarding is ready")

	sseURL := fmt.Sprintf("http://localhost:%d/flows/_stream", freePort)
	if err := util.ConsumeSSEStream(sseURL); err != nil {
		return fmt.Errorf("error consuming SSE stream: %w", err)
	}

	// Keep the program running until interrupted
	<-stopChan
	return nil
}
