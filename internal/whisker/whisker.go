package whisker

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/doucol/clyde/internal/catcher"
	"github.com/doucol/clyde/internal/flowdata"
	log "github.com/sirupsen/logrus"
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

	flowCatcher := func(data string) error {
		var fr flowdata.FlowResponse
		if err := json.Unmarshal([]byte(data), &fr); err != nil {
			return err
		}
		fd := &flowdata.FlowData{FlowResponse: fr}
		return fds.Add(fd)
	}

	dc := catcher.NewDataCatcher(ctx, CalicoNamespace, WhiskerContainer, "PORT", UrlPath, flowCatcher)

	// Go capture flows
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := dc.CatchDataFromSSEStream(); err != nil {
			panic(err)
		}
		log.Debug("exiting data capture routine")
	}()

	// Go run the flow watcher app
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := flowApp.Run(); err != nil {
			panic(err)
		}
		log.Debug("exiting flow watcher app")
	}()

	// Wait for both goroutines to finish
	wg.Wait()
	return nil
}
