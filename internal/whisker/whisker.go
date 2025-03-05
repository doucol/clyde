package whisker

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/util"
	"k8s.io/apimachinery/pkg/util/runtime"
)

const (
	CalicoNamespace  = "calico-system"
	WhiskerContainer = "whisker-backend"
	UrlPath          = "/flows?watch=true"
)

func WatchFlows(ctx context.Context) error {
	wg := sync.WaitGroup{}
	fds, err := flowdata.NewFlowDataStore()
	if err != nil {
		return err
	}
	defer fds.Close()

	flowApp := NewFlowApp(ctx, fds)

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

	dc := util.NewDataCatcher(ctx, CalicoNamespace, WhiskerContainer, "PORT", UrlPath, flowCatcher)

	// Go capture flows
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.HandleError(dc.CatchDataFromSSEStream())
	}()

	// Go run the flow watcher app
	wg.Add(1)
	go func() {
		defer wg.Done()
		runtime.HandleError(flowApp.Run())
	}()

	// Wait for both goroutines to finish
	wg.Wait()
	return nil
}
