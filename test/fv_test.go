package test

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/util"
	"github.com/doucol/clyde/internal/whisker"
	"github.com/doucol/clyde/test/mock"
	"github.com/sirupsen/logrus"
)

func TestFlowData(t *testing.T) {
	logrus.SetLevel(logrus.InfoLevel)
	t.Log("Starting Functional Validation (FV) testing...")
	ctx, cancel := context.WithCancel(context.Background())
	wgMock := sync.WaitGroup{}
	wgWhisker := sync.WaitGroup{}

	config := mock.DefaultConfig()
	config.AutoBroadcast = false
	server := mock.NewSSEServer(config)

	if err := flowdata.Clear(); err != nil {
		t.Fatalf("Failed to clear flow data: %v", err)
	}
	mockReady := make(chan bool)
	whiskerReady := make(chan bool)

	wgMock.Add(1)
	go func() {
		defer wgMock.Done()
		if err := server.Start(mockReady); err != nil {
			log.Fatalf("Failed to start SSE server: %v", err)
		}
	}()

	t.Log("Waiting for mock server to be ready")
	if err := util.ChanWaitTimeout(mockReady, 2, nil); err != nil {
		t.Fatalf("Failed to wait for mock server ready: %v", err)
	} else {
		t.Log("Mock server is ready")
	}

	wgWhisker.Add(1)
	go func() {
		defer wgWhisker.Done()
		wc := whisker.DefaultConfig()
		wc.TerminalUI = false
		wc.URL = server.URL()
		if err := whisker.WatchFlows(ctx, wc, whiskerReady); err != nil {
			log.Fatalf("Failed to watch flows: %v", err)
		}
	}()

	t.Log("Waiting for whisker to be ready")
	if err := util.ChanWaitTimeout(whiskerReady, 2, nil); err != nil {
		t.Fatalf("Failed to wait for whisker to be ready: %v", err)
	} else {
		t.Log("Whisker is ready")
	}

	if err := util.ChanWaitTimeout(server.Connect, 2, nil); err != nil {
		t.Fatalf("Failed to wait for SSE server connection: %v", err)
	} else {
		t.Logf("Whisker has connected to SSE server at %s", server.URL())
	}

	// now that we have our mock SSE server up & whisker connected, let's
	// broadcast some flow pairs
	pairs := server.GenerateFlowPairs()
	server.BroadcastFlowPairs(pairs)

	t.Logf("Sleeping to allow data processing")
	time.Sleep(2 * time.Second)

	t.Logf("Cancelling context to stop data processing")
	cancel()
	t.Logf("Waiting for whisker to finish")
	wgWhisker.Wait()

	// We have to wait for whisker to shutdown before we can open the
	// flowdata store.
	fds, err := flowdata.NewFlowDataStore()
	if err != nil {
		t.Fatalf("Failed to create flow data store: %v", err)
	}
	defer fds.Close()

	fss := fds.GetFlowSums(flowdata.FilterAttributes{})
	if len(fss) != len(pairs) {
		t.Fatalf("Flow sum count %d does not match test count %d", len(fss), len(pairs))
	}
	for _, fs := range fss {
		if p, exists := pairs[fs.Key]; exists {
			if uint64(p.SrcFlow.BytesIn) != fs.SourceBytesIn {
				t.Errorf("Flow sum does match broadcast pair for %s", fs.Key)
			}
			if uint64(p.DstFlow.BytesIn) != fs.DestBytesIn {
				t.Errorf("Flow sum does match broadcast pair for %s", fs.Key)
			}
			if uint64(p.SrcFlow.PacketsIn) != fs.SourcePacketsIn {
				t.Errorf("Flow sum does match broadcast pair for %s", fs.Key)
			}
			if uint64(p.DstFlow.PacketsIn) != fs.DestPacketsIn {
				t.Errorf("Flow sum does match broadcast pair for %s", fs.Key)
			}
			if uint64(p.SrcFlow.BytesOut) != fs.SourceBytesOut {
				t.Errorf("Flow sum does match broadcast pair for %s", fs.Key)
			}
			if uint64(p.DstFlow.BytesOut) != fs.DestBytesOut {
				t.Errorf("Flow sum does match broadcast pair for %s", fs.Key)
			}
			if uint64(p.SrcFlow.PacketsOut) != fs.SourcePacketsOut {
				t.Errorf("Flow sum does match broadcast pair for %s", fs.Key)
			}
			if uint64(p.DstFlow.PacketsOut) != fs.DestPacketsOut {
				t.Errorf("Flow sum does match broadcast pair for %s", fs.Key)
			}
		} else {
			t.Errorf("Flow sum %s not found in broadcast pairs", fs.Key)
		}
	}

	if err := server.Close(); err != nil {
		t.Fatalf("Failed to close SSE server: %v", err)
	}
	wgMock.Wait()
}
