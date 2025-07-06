// Package whisker provides functionality to watch and display network flows
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
	"github.com/doucol/clyde/internal/util"
	"github.com/sirupsen/logrus"
)

type WhiskerConfig struct {
	TerminalUI       bool
	CalicoNamespace  string
	WhiskerContainer string
	URL              string
	URLPath          string
	RecoverFunc      func()
	CatcherFunc      catcher.CatcherFunc
}

func DefaultConfig() *WhiskerConfig {
	return &WhiskerConfig{
		TerminalUI:       os.Getenv("NOTUI") == "",
		CalicoNamespace:  "calico-system",
		WhiskerContainer: "whisker-backend",
		URLPath:          "/flows?watch=true",
		RecoverFunc:      nil,
		CatcherFunc:      nil,
		URL:              "",
	}
}

func WatchFlows(ctx context.Context, cfg *WhiskerConfig, whiskerReady chan bool) error {
	wg := &sync.WaitGroup{}
	fds, err := flowdata.NewFlowDataStore()
	if err != nil {
		return err
	}
	defer fds.Close()

	flowCache := flowcache.NewFlowCache(ctx, fds)
	flowApp := tui.NewFlowApp(fds, flowCache)

	recoverFunc := cfg.RecoverFunc
	if recoverFunc == nil {
		recoverFunc = func() {
			if err := recover(); err != nil {
				flowApp.Stop()
				panic(err)
			}
		}
	}

	// Go capture flows
	fds.Run(recoverFunc)

	flowCatcher := cfg.CatcherFunc
	if flowCatcher == nil {
		flowCatcher = func(data string) error {
			var fr flowdata.FlowResponse
			if err := json.Unmarshal([]byte(data), &fr); err != nil {
				logrus.Panicf("error unmarshalling flow data: %v", err)
			}
			fd := &flowdata.FlowData{FlowResponse: fr}
			fds.AddFlow(fd)
			return nil
		}
	}

	sseReady := make(chan bool, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer flowApp.Stop()
		defer recoverFunc()
		tock := time.Tick(2 * time.Second)
		var lastError error
		for {
			dc := catcher.NewDataCatcher(cfg.CalicoNamespace, cfg.WhiskerContainer, cfg.URLPath, flowCatcher, recoverFunc)
			dc.URLFull = cfg.URL
			if err := dc.CatchServerSentEvents(ctx, sseReady); err != nil {
				// Don't keep logging the same error
				if !errors.Is(err, lastError) {
					lastError = err
					logrus.Debugf("error in flow catcher: %s", err.Error())
				}
			}
			select {
			case <-ctx.Done():
				logrus.Debug("exiting flow catcher routine: done signal received")
				return
			case <-tock:
				select {
				case <-ctx.Done():
					// If we are already done, don't restart the flow catcher
					return
				default:
					logrus.Debug("restarting the flow catcher")
					continue
				}
			}
		}
	}()

	// Go run the flow watcher TUI app
	if cfg.TerminalUI {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer flowApp.Stop()
			defer recoverFunc()
			if err := flowApp.Run(ctx); err != nil {
				logrus.Panicf("error running flow app: %v", err)
			}
			logrus.Debug("exiting flow watcher tui app")
		}()
	}

	if _, err := util.ChanWaitTimeout(sseReady, 2, whiskerReady); err != nil {
		logrus.Panicf("error waiting for sse consumer to be ready: %v", err)
	} else {
		logrus.Debug("Whisker is running and waiting for goroutines to exit")
	}

	// Wait for both goroutines to finish
	wg.Wait()
	logrus.Debug("exiting watch flows")
	return nil
}
