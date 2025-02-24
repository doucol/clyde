package whisker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/alphadose/haxmap"
	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/util"
	"github.com/rivo/tview"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type FlowResponse struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Action          string    `json:"action"`
	SourceName      string    `json:"source_name"`
	SourceNamespace string    `json:"source_namespace"`
	SourceLabels    string    `json:"source_labels"`
	DestName        string    `json:"dest_name"`
	DestNamespace   string    `json:"dest_namespace"`
	DestLabels      string    `json:"dest_labels"`
	Protocol        string    `json:"protocol"`
	DestPort        int64     `json:"dest_port"`
	Reporter        string    `json:"reporter"`
	PacketsIn       int64     `json:"packets_in"`
	PacketsOut      int64     `json:"packets_out"`
	BytesIn         int64     `json:"bytes_in"`
	BytesOut        int64     `json:"bytes_out"`
}

const (
	CalicoNamespace             = "calico-system"
	WhiskerBackendContainerName = "whisker-backend"
)

var (
	app   = tview.NewApplication()
	flows = haxmap.New[int, FlowResponse]()
)

type flowTable struct {
	tview.TableContentReadOnly
}

func (ft *flowTable) GetCell(row, column int) *tview.TableCell {
	if column == 0 {
		return tview.NewTableCell(fmt.Sprintf("%d", row))
	}
	if val, ok := flows.Get(row); ok {
		return tview.NewTableCell(fmt.Sprintf("%s|%s|%s|%s|%s", val.Action, val.SourceNamespace, val.SourceName, val.DestNamespace, val.DestName))
	}
	panic("invalid cell")
}

func (ft *flowTable) GetRowCount() int {
	return int(flows.Len())
}

func (ft *flowTable) GetColumnCount() int {
	return 2
}

func WatchFlows(ctx context.Context) error {
	go func() {
		if err := streamFlows(ctx); err != nil {
			fmt.Printf("Error streaming flows: %v\n", err)
		}
	}()
	flex := tview.NewFlex()
	flex.SetBorder(true).SetTitle("Calico Flows")
	table := tview.NewTable().SetBorders(false).SetSelectable(true, false).SetContent(&flowTable{})
	flex.AddItem(table, 0, 1, true)
	if err := app.SetRoot(flex, true).Run(); err != nil {
		return err
	}
	return nil
}

func flowCatcher(data string) {
	// fmt.Printf("Received event data: %s\n", data)
	var fr FlowResponse
	if err := json.Unmarshal([]byte(data), &fr); err != nil {
		panic(err)
	}
	flows.Set(int(flows.Len()), fr)
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
