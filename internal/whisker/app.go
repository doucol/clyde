package whisker

import (
	"context"
	"fmt"

	"github.com/doucol/clyde/internal/cmdContext"
	"github.com/doucol/clyde/internal/flowdata"
	tcell "github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FlowApp struct {
	ctx context.Context
	app *tview.Application
	fds *flowdata.FlowDataStore
}

func NewFlowApp(ctx context.Context, fds *flowdata.FlowDataStore) *FlowApp {
	return &FlowApp{ctx, tview.NewApplication(), fds}
}

func (fa *FlowApp) Run() error {
	cc := cmdContext.CmdContextFromContext(fa.ctx)
	fa.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlC {
			fa.app.Stop()
			cc.Cancel()
			return nil
		}
		return event
	})

	if err := fa.ViewSummary().Run(); err != nil {
		return err
	}
	return nil
}

func (fa *FlowApp) ViewSummary() *tview.Application {
	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowSumTable{fds: fa.fds}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			row, _ := tableData.GetSelection()
			if row > 0 {
				sns := tableData.GetCell(row, 0).Text
				sn := tableData.GetCell(row, 1).Text
				dns := tableData.GetCell(row, 2).Text
				dn := tableData.GetCell(row, 3).Text
				proto := tableData.GetCell(row, 4).Text
				port := tableData.GetCell(row, 5).Text
				key := fmt.Sprintf("%s|%s|%s|%s|%s|%s", sns, sn, dns, dn, proto, port)
				fa.ViewDetail(key)
				return nil
			}
		}
		return event
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary")
	flex.AddItem(tableData, 0, 1, true)
	fa.app.SetRoot(flex, true)
	return fa.app
}

func (fa *FlowApp) ViewDetail(key string) *tview.Application {
	tableDataHeader := tview.NewTable().SetBorders(true).SetSelectable(true, false).
		SetContent(&flowDetailTableHeader{fds: fa.fds, key: key})

	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
		SetContent(&flowDetailTable{fds: fa.fds, key: key}).SetFixed(1, 0)

	tableData.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			fa.ViewSummary()
			return nil
		}
		return event
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Detail")
	flex.AddItem(tableDataHeader, 6, 1, false)
	flex.AddItem(tableData, 0, 1, true)
	fa.app.SetRoot(flex, true)
	return fa.app
}

// func WatchFlows(ctx context.Context) error {
// 	var err error
// 	fds, err = flowdata.NewFlowDataStore()
// 	if err != nil {
// 		return err
// 	}
// 	defer fds.Close()
//
// 	go func() {
// 		if err := streamFlows(ctx); err != nil {
// 			panic(fmt.Sprintf("Error streaming flows: %v\n", err))
// 		}
// 	}()
//
// 	tableData := tview.NewTable().SetBorders(false).SetSelectable(true, false).
// 		SetContent(&flowTable{}).SetFixed(1, 0)
//
// 	flex := tview.NewFlex()
// 	flex.SetDirection(tview.FlexRow).SetBorder(true).SetTitle("Calico Flow Summary")
// 	flex.AddItem(tableData, 0, 1, true)
//
// 	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
// 		if event.Key() == tcell.KeyCtrlC {
// 			app.Stop()
// 			return nil
// 		}
// 		return event
// 	})
//
// 	if err := app.SetRoot(flex, true).Run(); err != nil {
// 		return err
// 	}
// 	return nil
// }
//
// func flowCatcher(data string) {
// 	var fr flowdata.FlowResponse
// 	if err := json.Unmarshal([]byte(data), &fr); err != nil {
// 		panic(err)
// 	}
// 	fd := &flowdata.FlowData{FlowResponse: fr}
// 	err := fds.Add(fd)
// 	if err != nil {
// 		panic(err)
// 	}
// 	app.Draw()
// }
//
// func streamFlows(ctx context.Context) error {
// 	config := cmdContext.K8sConfigFromContext(ctx)
//
// 	// URL for the portforward endpoint on the pod
// 	podName, port, err := util.GetPodAndEnvVarWithContainerName(ctx, CalicoNamespace, WhiskerBackendContainerName, "PORT")
// 	if err != nil {
// 		return err
// 	}
//
// 	apiURL, _ := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/%s/portforward", config.Host, CalicoNamespace, podName))
//
// 	// Dialer for establishing the connection
// 	transport, upgrader, err := spdy.RoundTripperFor(config)
// 	if err != nil {
// 		return err
// 	}
// 	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, apiURL)
//
// 	// Port mappings (local port: pod port)
// 	freePort, err := util.GetFreePort()
// 	if err != nil {
// 		return err
// 	}
// 	ports := []string{fmt.Sprintf("%d:%s", freePort, port)}
//
// 	// Channels for signaling
// 	stopChan := make(chan struct{}, 1)
// 	readyChan := make(chan struct{})
//
// 	pf, err := portforward.New(dialer, ports, stopChan, readyChan, io.Discard, io.Discard)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Start the port forwarding
// 	go func() {
// 		if err := pf.ForwardPorts(); err != nil {
// 			panic(err)
// 		}
// 	}()
//
// 	// Wait for the port forwarding to be ready
// 	<-readyChan
//
// 	go func() {
// 		<-ctx.Done()
// 		pf.Close()
// 	}()
//
// 	sseURL := fmt.Sprintf("http://localhost:%d/flows?watch=true", freePort)
// 	if err := util.ConsumeSSEStream(ctx, sseURL, flowCatcher); err != nil {
// 		return fmt.Errorf("error consuming SSE stream: %w", err)
// 	}
//
// 	// Keep the program running until interrupted
// 	<-stopChan
// 	return nil
// }
//
