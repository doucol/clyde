package whisker

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/util"
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
	}

	dc := util.NewDataCatcher(ctx, CalicoNamespace, WhiskerContainer, "PORT", UrlPath, flowCatcher)

	// Go capture flows
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := dc.CatchDataFromSSEStream(); err != nil {
			panic(err)
		}
	}()

	// Go run the flow watcher app
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := flowApp.Run(); err != nil {
			panic(err)
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()
	return nil
}
