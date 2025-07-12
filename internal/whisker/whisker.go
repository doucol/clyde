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
	RateCalcWindow   int
	RateCalcInterval int
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
		RateCalcWindow:   60,
		RateCalcInterval: 5,
	}
}

type Whisker struct {
	cfg *WhiskerConfig
	fds *flowdata.FlowDataStore
}

func New(cfg *WhiskerConfig) *Whisker {
	return &Whisker{cfg: cfg}
}

func (w *Whisker) Config() *WhiskerConfig {
	return w.cfg
}

func (w *Whisker) FlowAdded() chan flowdata.Flower {
	return w.fds.FlowAdded()
}

func (w *Whisker) FlowSumAdded() chan flowdata.Flower {
	return w.fds.FlowSumAdded()
}

func (w *Whisker) FlowSumsUpdated() chan flowdata.Flower {
	return w.fds.FlowSumsUpdated()
}

func (w *Whisker) FlowRatesUpdated() chan flowdata.Flower {
	return w.fds.FlowRatesUpdated()
}

func (w *Whisker) WatchFlows(ctx context.Context, whiskerReady chan bool) error {
	var err error
	wg := &sync.WaitGroup{}
	w.fds, err = flowdata.NewFlowDataStore()
	if err != nil {
		return err
	}
	defer w.fds.Close()
	w.fds.RateCalcWindow = w.cfg.RateCalcWindow
	w.fds.RateCalcInterval = w.cfg.RateCalcInterval

	flowCache := flowcache.NewFlowCache(ctx, w.fds)
	flowApp := tui.NewFlowApp(w.fds, flowCache)

	recoverFunc := w.cfg.RecoverFunc
	if recoverFunc == nil {
		recoverFunc = func() {
			if err := recover(); err != nil {
				flowApp.Stop()
				panic(err)
			}
		}
	}

	// Go capture flows
	w.fds.Run(recoverFunc)

	flowCatcher := w.cfg.CatcherFunc
	if flowCatcher == nil {
		flowCatcher = func(data string) error {
			var fr flowdata.FlowResponse
			if err := json.Unmarshal([]byte(data), &fr); err != nil {
				logrus.Panicf("error unmarshalling flow data: %v", err)
			}
			fd := &flowdata.FlowData{FlowResponse: fr}
			w.fds.AddFlow(fd)
			return nil
		}
	}

	sseReady := make(chan bool)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer flowApp.Stop()
		defer recoverFunc()
		tock := time.Tick(2 * time.Second)
		var lastError error
		for {
			dc := catcher.NewDataCatcher(w.cfg.CalicoNamespace, w.cfg.WhiskerContainer, w.cfg.URLPath, flowCatcher, recoverFunc)
			dc.URLFull = w.cfg.URL
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
	if w.cfg.TerminalUI {
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
		logrus.WithError(err).Panic("error waiting for sse consumer to be ready")
	} else {
		logrus.Debug("Whisker is running and waiting for an exit signal")
	}

	// Wait for both goroutines to finish
	wg.Wait()
	logrus.Debug("exiting watch flows")
	return nil
}
