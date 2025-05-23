package whisker

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/doucol/clyde/internal/catcher"
	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/tui"
	log "github.com/sirupsen/logrus"
)

const (
	CalicoNamespace  = "calico-system"
	WhiskerContainer = "whisker-backend"
	UrlPath          = "/flows?watch=true"
)

var NOTUI = os.Getenv("NOTUI") != ""

func WatchFlows(ctx context.Context) error {
	// cctx := cmdCtx.CmdContextFromContext(ctx)
	wg := sync.WaitGroup{}
	fds, err := flowdata.NewFlowDataStore()
	if err != nil {
		return err
	}
	defer fds.Close()

	flowCache := flowcache.NewFlowCache(ctx, fds)
	flowApp := tui.NewFlowApp(fds, flowCache)

	recoverFunc := func() {
		if err := recover(); err != nil {
			flowApp.Stop()
			panic(err)
		}
	}

	// Go capture flows
	fds.Run(recoverFunc)

	flowCatcher := func(data string) error {
		var fr flowdata.FlowResponse
		if err := json.Unmarshal([]byte(data), &fr); err != nil {
			log.Panicf("error unmarshalling flow data: %v", err)
		}
		fd := &flowdata.FlowData{FlowResponse: fr}
		fds.AddFlow(fd)
		return nil
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer flowApp.Stop()
		defer recoverFunc()
		ticker := time.Tick(2 * time.Second)
		var lastError error
		for {
			dc := catcher.NewDataCatcher(CalicoNamespace, WhiskerContainer, UrlPath, flowCatcher, recoverFunc)
			if err := dc.CatchDataFromSSEStream(ctx); err != nil {
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
				select {
				case <-ctx.Done():
					// If we are already done, don't restart the flow catcher
					return
				default:
					log.Debug("restarting the flow catcher")
					continue
				}
			}
		}
	}()

	// Go run the flow watcher app
	if !NOTUI {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer flowApp.Stop()
			defer recoverFunc()
			if err := flowApp.Run(ctx); err != nil {
				log.Panicf("error running flow app: %v", err)
			}
			log.Debug("exiting flow watcher tui app")
		}()
	}
	// Wait for both goroutines to finish
	wg.Wait()
	log.Debug("exiting watch flows")
	return nil
}
