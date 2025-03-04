package whisker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/util"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	CalicoNamespace             = "calico-system"
	WhiskerBackendContainerName = "whisker-backend"
)

func WatchFlows(ctx context.Context) error {
	fds, err := flowdata.NewFlowDataStore()
	if err != nil {
		return err
	}
	defer fds.Close()

	flowApp := NewFlowApp(fds)

	flowCatcher := func(data string) {
		var fr flowdata.FlowResponse
		if err := json.Unmarshal([]byte(data), &fr); err != nil {
			panic(err)
		}
		fd := &flowdata.FlowData{FlowResponse: fr}
		err := fds.Add(fd)
		if err != nil {
			panic(err)
		}
		flowApp.app.Draw()
	}

	go func() {
		if err := streamFlows(ctx, flowCatcher); err != nil {
			panic(fmt.Sprintf("Error streaming flows: %v\n", err))
		}
	}()

	if err := flowApp.Run(); err != nil {
		return err
	}
	return nil
}

func streamFlows(ctx context.Context, flowCatcher func(data string)) error {
	config := cmdContext.K8sConfigFromContext(ctx)

	// URL for the portforward endpoint on the pod
	podName, port, err := util.GetPodAndEnvVarWithContainerName(ctx, CalicoNamespace, WhiskerBackendContainerName, "PORT")
	if err != nil {
		return err
	}

	apiURL, _ := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", config.Host, CalicoNamespace, podName))

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

	pf, err := portforward.New(dialer, ports, stopChan, readyChan, io.Discard, io.Discard)
	if err != nil {
		return err
	}

	// Start the port forwarding
	go func() {
		if err := pf.ForwardPorts(); err != nil {
			panic(err)
		}
	}()

	// Wait for the port forwarding to be ready
	<-readyChan

	go func() {
		<-ctx.Done()
		pf.Close()
	}()

	sseURL := fmt.Sprintf("http://localhost:%d/flows?watch=true", freePort)
	if err := util.ConsumeSSEStream(ctx, sseURL, flowCatcher); err != nil {
		return fmt.Errorf("error consuming SSE stream: %w", err)
	}

	// Keep the program running until interrupted
	<-stopChan
	return nil
}
