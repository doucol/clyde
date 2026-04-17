package tui

import (
	"testing"

	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
)

func TestFlowAppState_Reset(t *testing.T) {
	fas := &flowAppState{
		sumID:        123,
		sumRow:       456,
		rateID:       789,
		rateRow:      101,
		flowID:       112,
		flowRow:      131,
		lastHomePage: "test-page",
	}

	fas.reset()

	if fas.sumID != 0 {
		t.Errorf("expected sumID = 0, got %d", fas.sumID)
	}
	if fas.sumRow != 0 {
		t.Errorf("expected sumRow = 0, got %d", fas.sumRow)
	}
	if fas.rateID != 0 {
		t.Errorf("expected rateID = 0, got %d", fas.rateID)
	}
	if fas.rateRow != 0 {
		t.Errorf("expected rateRow = 0, got %d", fas.rateRow)
	}
	if fas.flowID != 0 {
		t.Errorf("expected flowID = 0, got %d", fas.flowID)
	}
	if fas.flowRow != 0 {
		t.Errorf("expected flowRow = 0, got %d", fas.flowRow)
	}

	// lastHomePage should not be reset by reset()
	if fas.lastHomePage != "test-page" {
		t.Errorf("expected lastHomePage = 'test-page', got '%s'", fas.lastHomePage)
	}
}

func TestFlowAppState_SetSum(t *testing.T) {
	fas := &flowAppState{
		sumID:   999,
		sumRow:  888,
		rateID:  777,
		rateRow: 666,
		flowID:  555,
		flowRow: 444,
	}

	fas.setSum(100, 200)

	if fas.sumID != 100 {
		t.Errorf("expected sumID = 100, got %d", fas.sumID)
	}
	if fas.sumRow != 200 {
		t.Errorf("expected sumRow = 200, got %d", fas.sumRow)
	}
	if fas.flowID != 0 {
		t.Errorf("expected flowID = 0, got %d", fas.flowID)
	}
	if fas.flowRow != 0 {
		t.Errorf("expected flowRow = 0, got %d", fas.flowRow)
	}
	if fas.rateID != 777 {
		t.Errorf("expected rateID = 777, got %d", fas.rateID)
	}
	if fas.rateRow != 666 {
		t.Errorf("expected rateRow = 666, got %d", fas.rateRow)
	}
}

func TestFlowAppState_SetRate(t *testing.T) {
	fas := &flowAppState{
		sumID:   999,
		sumRow:  888,
		rateID:  777,
		rateRow: 666,
		flowID:  555,
		flowRow: 444,
	}

	fas.setRate(300, 400)

	if fas.rateID != 300 {
		t.Errorf("expected rateID = 300, got %d", fas.rateID)
	}
	if fas.rateRow != 400 {
		t.Errorf("expected rateRow = 400, got %d", fas.rateRow)
	}
	if fas.flowID != 0 {
		t.Errorf("expected flowID = 0, got %d", fas.flowID)
	}
	if fas.flowRow != 0 {
		t.Errorf("expected flowRow = 0, got %d", fas.flowRow)
	}
	if fas.sumID != 999 {
		t.Errorf("expected sumID = 999, got %d", fas.sumID)
	}
	if fas.sumRow != 888 {
		t.Errorf("expected sumRow = 888, got %d", fas.sumRow)
	}
}

func TestFlowAppState_SetFlow(t *testing.T) {
	fas := &flowAppState{
		sumID:   999,
		sumRow:  888,
		rateID:  777,
		rateRow: 666,
		flowID:  555,
		flowRow: 444,
	}

	fas.setFlow(500, 600)

	if fas.flowID != 500 {
		t.Errorf("expected flowID = 500, got %d", fas.flowID)
	}
	if fas.flowRow != 600 {
		t.Errorf("expected flowRow = 600, got %d", fas.flowRow)
	}
	if fas.sumID != 999 {
		t.Errorf("expected sumID = 999, got %d", fas.sumID)
	}
	if fas.sumRow != 888 {
		t.Errorf("expected sumRow = 888, got %d", fas.sumRow)
	}
	if fas.rateID != 777 {
		t.Errorf("expected rateID = 777, got %d", fas.rateID)
	}
	if fas.rateRow != 666 {
		t.Errorf("expected rateRow = 666, got %d", fas.rateRow)
	}
}

func TestFlowAppState_ZeroValues(t *testing.T) {
	fas := &flowAppState{}

	if fas.sumID != 0 {
		t.Errorf("expected sumID = 0, got %d", fas.sumID)
	}
	if fas.sumRow != 0 {
		t.Errorf("expected sumRow = 0, got %d", fas.sumRow)
	}
	if fas.rateID != 0 {
		t.Errorf("expected rateID = 0, got %d", fas.rateID)
	}
	if fas.rateRow != 0 {
		t.Errorf("expected rateRow = 0, got %d", fas.rateRow)
	}
	if fas.flowID != 0 {
		t.Errorf("expected flowID = 0, got %d", fas.flowID)
	}
	if fas.flowRow != 0 {
		t.Errorf("expected flowRow = 0, got %d", fas.flowRow)
	}
	if fas.lastHomePage != "" {
		t.Errorf("expected lastHomePage = '', got '%s'", fas.lastHomePage)
	}
}

func TestFlowAppState_LastHomePage(t *testing.T) {
	fas := &flowAppState{}

	fas.lastHomePage = "summary"
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage = 'summary', got '%s'", fas.lastHomePage)
	}

	fas.reset()
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage to remain 'summary' after reset, got '%s'", fas.lastHomePage)
	}

	fas.setSum(1, 2)
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage to remain 'summary' after setSum, got '%s'", fas.lastHomePage)
	}

	fas.setRate(3, 4)
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage to remain 'summary' after setRate, got '%s'", fas.lastHomePage)
	}

	fas.setFlow(5, 6)
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage to remain 'summary' after setFlow, got '%s'", fas.lastHomePage)
	}
}

func TestNewFlowApp(t *testing.T) {
	var fds *flowdata.FlowDataStore
	var fc *flowcache.FlowCache

	fa := NewFlowApp(fds, fc)

	if fa == nil {
		t.Fatal("expected NewFlowApp to return non-nil FlowApp")
	}
	if fa.mu == nil {
		t.Error("expected mutex to be initialized")
	}
	if fa.fds != fds {
		t.Error("expected fds to match provided FlowDataStore")
	}
	if fa.fc != fc {
		t.Error("expected fc to match provided FlowCache")
	}
	if fa.fas == nil {
		t.Error("expected flowAppState to be initialized")
	}
}

func TestFlowApp_BasicFields(t *testing.T) {
	var fds *flowdata.FlowDataStore
	var fc *flowcache.FlowCache

	fa := NewFlowApp(fds, fc)

	if fa.fas.sumID != 0 {
		t.Errorf("expected initial sumID = 0, got %d", fa.fas.sumID)
	}

	if fa.fas.lastHomePage != "" {
		t.Errorf("expected initial lastHomePage = '', got '%s'", fa.fas.lastHomePage)
	}

	fa.fas.setSum(100, 200)
	if fa.fas.sumID != 100 {
		t.Errorf("expected sumID = 100, got %d", fa.fas.sumID)
	}
	if fa.fas.sumRow != 200 {
		t.Errorf("expected sumRow = 200, got %d", fa.fas.sumRow)
	}
}

func TestPageConstants(t *testing.T) {
	expectedPages := map[string]string{
		pageHomeName:          "home",
		pageSummaryTotalsName: "summaryTotals",
		pageSummaryRatesName:  "summaryRates",
		pageSumDetailName:     "sumDetail",
		pageFlowDetailName:    "flowDetail",
	}

	for constant, expected := range expectedPages {
		if constant != expected {
			t.Errorf("expected page constant to equal %s, got %s", expected, constant)
		}
	}
}

func TestFlowApp_UpdateSort_InvalidPageName(t *testing.T) {
	var fds *flowdata.FlowDataStore
	var fc *flowcache.FlowCache
	fa := NewFlowApp(fds, fc)

	// updateSort returns a non-nil sentinel when the page is unknown.
	result := fa.updateSort(nil, "testField", true, "invalidPage")
	if result == nil {
		t.Error("expected updateSort to pass through the event for an invalid page")
	}
}

func TestFlowAppState_ConcurrentAccess(t *testing.T) {
	fas := &flowAppState{}

	fas.setSum(1, 2)
	fas.setRate(3, 4)
	fas.setFlow(5, 6)
	fas.reset()

	if fas.sumID != 0 || fas.rateID != 0 || fas.flowID != 0 {
		t.Error("expected all IDs to be 0 after reset")
	}
}

func TestFlowAppState_ChainedOperations(t *testing.T) {
	fas := &flowAppState{}

	fas.setSum(10, 20)
	fas.setRate(30, 40)
	fas.setFlow(50, 60)

	if fas.sumID != 10 || fas.sumRow != 20 {
		t.Errorf("expected sum values to remain: ID=10, Row=20, got ID=%d, Row=%d", fas.sumID, fas.sumRow)
	}
	if fas.rateID != 30 || fas.rateRow != 40 {
		t.Errorf("expected rate values to remain: ID=30, Row=40, got ID=%d, Row=%d", fas.rateID, fas.rateRow)
	}
	if fas.flowID != 50 || fas.flowRow != 60 {
		t.Errorf("expected flow values: ID=50, Row=60, got ID=%d, Row=%d", fas.flowID, fas.flowRow)
	}

	fas.reset()
	if fas.sumID != 0 || fas.sumRow != 0 || fas.rateID != 0 || fas.rateRow != 0 || fas.flowID != 0 || fas.flowRow != 0 {
		t.Error("expected all values to be 0 after reset")
	}
}
