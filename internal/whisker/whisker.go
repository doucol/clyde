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
	"github.com/rivo/tview"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	CalicoNamespace             = "calico-system"
	WhiskerBackendContainerName = "whisker-backend"
)

var (
	app = tview.NewApplication()
	fds *flowdata.FlowDataStore
)

func WatchFlows(ctx context.Context) error {
	var err error
	fds, err = flowdata.NewFlowDataStore()
	if err != nil {
		return err
	}
	defer fds.Close()
	go func() {
		if err := streamFlows(ctx); err != nil {
			fmt.Printf("Error streaming flows: %v\n", err)
		}
	}()
	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowTable{}).SetFixed(1, 0)

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flows")
	flex.AddItem(tableData, 0, 1, true)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		return err
	}
	return nil
}

func flowCatcher(data string) {
	// fmt.Printf("Received event data: %s\n", data)
	var fr flowdata.FlowResponse
	if err := json.Unmarshal([]byte(data), &fr); err != nil {
		panic(err)
	}
	fd := &flowdata.FlowData{FlowResponse: fr}
	err := fds.Add(fd)
	if err != nil {
		panic(err)
	}
	app.Draw()
}

func streamFlows(ctx context.Context) error {
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

	// Create the port forwarder
	// forwarder, err := portforward.New(dialer, ports, stopChan, readyChan, os.Stdout, os.Stderr)
	forwarder, err := portforward.New(dialer, ports, stopChan, readyChan, io.Discard, io.Discard)
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

	sseURL := fmt.Sprintf("http://localhost:%d/flows?watch=true", freePort)
	if err := util.ConsumeSSEStream(sseURL, flowCatcher); err != nil {
		return fmt.Errorf("error consuming SSE stream: %w", err)
	}

	// Keep the program running until interrupted
	<-stopChan
	return nil
}
