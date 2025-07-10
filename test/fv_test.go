package test

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/util"
	"github.com/doucol/clyde/internal/whisker"
	"github.com/doucol/clyde/test/mock"
	"github.com/sirupsen/logrus"
)

func clearTestData(t *testing.T) {
	if err := flowdata.Clear(); err != nil {
		t.Fatalf("Failed to clear flow data: %v", err)
	}
}

func TestFlowData(t *testing.T) {
	t.Log("Starting Functional Validation (FV) testing...")
	logrus.SetLevel(logrus.InfoLevel)
	if err := os.Setenv("XDG_DATA_HOME", os.TempDir()); err != nil {
		t.Fatalf("Failed to set XDG_DATA_HOME: %v", err)
	}

	clearTestData(t)
	defer clearTestData(t)

	wgMock := sync.WaitGroup{}
	wgWhisker := sync.WaitGroup{}

	config := mock.DefaultConfig()
	config.AutoBroadcast = false
	server := mock.NewSSEServer(config)

	mockReady := make(chan bool)
	whiskerReady := make(chan bool)

	ctx, cancel := context.WithCancel(context.Background())

	wgMock.Add(1)
	go func() {
		defer wgMock.Done()
		if err := server.Start(mockReady); err != nil {
			log.Fatalf("Failed to start SSE server: %v", err)
		}
	}()

	t.Log("Waiting for mock server to be ready")
	if _, err := util.ChanWaitTimeout(mockReady, 2); err != nil {
		t.Fatalf("Failed to wait for mock server ready: %v", err)
	} else {
		t.Log("Mock server is ready")
	}

	var wh *whisker.Whisker

	wgWhisker.Add(1)
	go func() {
		defer wgWhisker.Done()
		wc := whisker.DefaultConfig()
		wc.TerminalUI = false
		wc.URL = server.URL()
		wh = whisker.New(wc)
		if err := wh.WatchFlows(ctx, whiskerReady); err != nil {
			log.Fatalf("Failed to watch flows: %v", err)
		}
	}()

	t.Log("Waiting for whisker to be ready")
	if _, err := util.ChanWaitTimeout(whiskerReady, 2); err != nil {
		t.Fatalf("Failed to wait for whisker to be ready: %v", err)
	} else {
		t.Log("Whisker is ready")
	}

	if _, err := util.ChanWaitTimeout(server.Connect, 2); err != nil {
		t.Fatalf("Failed to wait for SSE server connection: %v", err)
	} else {
		t.Logf("Whisker has connected to SSE server at %s", server.URL())
	}

	// now that we have our mock SSE server up & whisker connected, let's
	// broadcast some flow pairs
	pairs := server.GenerateFlowPairs()
	server.BroadcastFlowPairs(pairs)

	t.Logf("Sleeping to allow data processing and rate calculations")
	var flowSumCount int
	for range wh.FlowRatesUpdated() {
		flowSumCount++
		if flowSumCount >= len(pairs) {
			break
		}
	}
	// time.Sleep(8 * time.Second)

	t.Logf("Cancelling context to stop whisker")
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

	sumSrcMap := map[string]string{
		"StartTime":        "StartTime",
		"EndTime":          "EndTime",
		"Action":           "Action",
		"SourceName":       "SourceName",
		"SourceNamespace":  "SourceNamespace",
		"SourceLabels":     "SourceLabels",
		"Protocol":         "Protocol",
		"DestPort":         "DestPort",
		"SourceBytesIn":    "BytesIn",
		"SourcePacketsIn":  "PacketsIn",
		"SourceBytesOut":   "BytesOut",
		"SourcePacketsOut": "PacketsOut",
	}
	sumDstMap := map[string]string{
		"StartTime":      "StartTime",
		"EndTime":        "EndTime",
		"Action":         "Action",
		"DestName":       "DestName",
		"DestNamespace":  "DestNamespace",
		"DestLabels":     "DestLabels",
		"Protocol":       "Protocol",
		"DestPort":       "DestPort",
		"DestBytesIn":    "BytesIn",
		"DestPacketsIn":  "PacketsIn",
		"DestBytesOut":   "BytesOut",
		"DestPacketsOut": "PacketsOut",
	}

	dtlMap := map[string]string{
		"StartTime":       "StartTime",
		"EndTime":         "EndTime",
		"Action":          "Action",
		"SourceName":      "SourceName",
		"SourceNamespace": "SourceNamespace",
		"SourceLabels":    "SourceLabels",
		"DestName":        "DestName",
		"DestNamespace":   "DestNamespace",
		"DestLabels":      "DestLabels",
		"Protocol":        "Protocol",
		"DestPort":        "DestPort",
		"Reporter":        "Reporter",
		"BytesIn":         "BytesIn",
		"PacketsIn":       "PacketsIn",
		"BytesOut":        "BytesOut",
		"PacketsOut":      "PacketsOut",
	}

	for _, fs := range fss {
		p, exists := pairs[fs.Key]
		if !exists {
			t.Errorf("Flow sum %s not found in broadcast pairs", fs.Key)
			continue
		}

		// check flow sum attributes
		for k, v := range sumSrcMap {
			if !util.CompareStructFields(fs, p.SrcFlow, k, v) {
				t.Fail()
			}
		}
		for k, v := range sumDstMap {
			if !util.CompareStructFields(fs, p.DstFlow, k, v) {
				t.Fail()
			}
		}
		if fs.SourceReports != 1 {
			t.Errorf("Expected SourceReports to be 1, got %d", fs.SourceReports)
		}
		if fs.DestReports != 1 {
			t.Errorf("Expected DestReports to be 1, got %d", fs.DestReports)
		}

		// check flowsum rates
		sec := fs.EndTime.Sub(fs.StartTime).Seconds()
		if sec != 15 {
			t.Errorf("Expected flow sum duration to be 15 seconds, got %f", sec)
		}

		// Individual source rates (in/out)
		if fs.SourcePacketsInRate != float64(p.SrcFlow.PacketsIn)/sec {
			t.Errorf("Expected SourcePacketsInRate to be %f, got %f", float64(p.SrcFlow.PacketsIn)/sec, fs.SourcePacketsInRate)
		}
		if fs.SourcePacketsOutRate != float64(p.SrcFlow.PacketsOut)/sec {
			t.Errorf("Expected SourcePacketsOutRate to be %f, got %f", float64(p.SrcFlow.PacketsOut)/sec, fs.SourcePacketsOutRate)
		}
		if fs.SourceBytesInRate != float64(p.SrcFlow.BytesIn)/sec {
			t.Errorf("Expected SourceBytesInRate to be %f, got %f", float64(p.SrcFlow.BytesIn)/sec, fs.SourceBytesInRate)
		}
		if fs.SourceBytesOutRate != float64(p.SrcFlow.BytesOut)/sec {
			t.Errorf("Expected SourceBytesOutRate to be %f, got %f", float64(p.SrcFlow.BytesOut)/sec, fs.SourceBytesOutRate)
		}

		// Individual destination rates (in/out)
		if fs.DestPacketsInRate != float64(p.DstFlow.PacketsIn)/sec {
			t.Errorf("Expected DestPacketsInRate to be %f, got %f", float64(p.DstFlow.PacketsIn)/sec, fs.DestPacketsInRate)
		}
		if fs.DestPacketsOutRate != float64(p.DstFlow.PacketsOut)/sec {
			t.Errorf("Expected DestPacketsOutRate to be %f, got %f", float64(p.DstFlow.PacketsOut)/sec, fs.DestPacketsOutRate)
		}
		if fs.DestBytesInRate != float64(p.DstFlow.BytesIn)/sec {
			t.Errorf("Expected DestBytesInRate to be %f, got %f", float64(p.DstFlow.BytesIn)/sec, fs.DestBytesInRate)
		}
		if fs.DestBytesOutRate != float64(p.DstFlow.BytesOut)/sec {
			t.Errorf("Expected DestBytesOutRate to be %f, got %f", float64(p.DstFlow.BytesOut)/sec, fs.DestBytesOutRate)
		}

		// Total source and destination rates (in + out)
		if fs.SourceTotalPacketRate != float64(p.SrcFlow.PacketsIn+p.SrcFlow.PacketsOut)/sec {
			t.Errorf("Expected SourceTotalPacketRate to be %f, got %f", float64(p.SrcFlow.PacketsIn+p.SrcFlow.PacketsOut)/sec, fs.SourceTotalPacketRate)
		}
		if fs.SourceTotalByteRate != float64(p.SrcFlow.BytesIn+p.SrcFlow.BytesOut)/sec {
			t.Errorf("Expected SourceTotalByteRate to be %f, got %f", float64(p.SrcFlow.BytesIn+p.SrcFlow.BytesOut)/sec, fs.SourceTotalByteRate)
		}
		if fs.DestTotalPacketRate != float64(p.DstFlow.PacketsIn+p.DstFlow.PacketsOut)/sec {
			t.Errorf("Expected DestTotalPacketRate to be %f, got %f", float64(p.DstFlow.PacketsIn+p.DstFlow.PacketsOut)/sec, fs.DestTotalPacketRate)
		}
		if fs.DestTotalByteRate != float64(p.DstFlow.BytesIn+p.DstFlow.BytesOut)/sec {
			t.Errorf("Expected DestTotalByteRate to be %f, got %f", float64(p.DstFlow.BytesIn+p.DstFlow.BytesOut)/sec, fs.DestTotalByteRate)
		}

		// check flowdata attributes under the flowsum
		flowDataSlice := fds.GetFlowsBySumID(fs.ID, flowdata.FilterAttributes{})
		if len(flowDataSlice) != 2 {
			t.Errorf("Expected 2 flow data entries for sum ID %d, got %d", fs.ID, len(flowDataSlice))
		}
		for _, fd := range flowDataSlice {
			var f *flowdata.FlowResponse
			switch fd.Reporter {
			case "Src":
				f = p.SrcFlow
			case "Dst":
				f = p.DstFlow
			default:
				t.Errorf("Unknown reporter %s in flow data", fd.Reporter)
				continue
			}
			for k, v := range dtlMap {
				if !util.CompareStructFields(fd, f, k, v) {
					t.Fail()
				}
			}
		}
	}

	if err := server.Close(); err != nil {
		t.Fatalf("Failed to close SSE server: %v", err)
	}
	wgMock.Wait()
}
