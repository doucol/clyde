package whisker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

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
		fs, newSum, err := fds.Add(fd)
		if err != nil {
			panic(fmt.Errorf("error adding flow data: %v", err.Error()))
		}
		if newSum {
			log.Tracef("added flow data: new flow sum: %s", fs.Key)
		} else {
			log.Tracef("added flow data: existing flow sum: %s", fs.Key)
		}
		return nil
	}

	dc := catcher.NewDataCatcher(ctx, CalicoNamespace, WhiskerContainer, "PORT", UrlPath, flowCatcher)

	// Go capture flows
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer flowApp.Stop()
		ticker := time.Tick(2 * time.Second)
		var lastError error
		for {
			if err := dc.CatchDataFromSSEStream(); err != nil {
				// Don't keep logging the same error
				if !errors.Is(err, lastError) {
					lastError = err
					log.Debugf("error in flow catcher: %s", err.Error())
				}
			}
			select {
			case <-ctx.Done():
				log.Debug("exiting flow catcher routine: done signal received")
				return
			case <-ticker:
				log.Debug("restarting the flow catcher")
				continue
			}
		}
	}()

	// Go run the flow watcher app
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer flowApp.Stop()
		if err := flowApp.Run(); err != nil {
			panic(err)
		}
		log.Debug("exiting flow watcher tui app")
	}()

	// Wait for both goroutines to finish
	wg.Wait()
	log.Debug("exiting watch flows")
	return nil
}
