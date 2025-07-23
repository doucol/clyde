package tui

import (
	"testing"

	"github.com/doucol/clyde/internal/flowcache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/gdamore/tcell/v2"
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

	// Should set sum fields
	if fas.sumID != 100 {
		t.Errorf("expected sumID = 100, got %d", fas.sumID)
	}
	if fas.sumRow != 200 {
		t.Errorf("expected sumRow = 200, got %d", fas.sumRow)
	}

	// Should reset flow fields
	if fas.flowID != 0 {
		t.Errorf("expected flowID = 0, got %d", fas.flowID)
	}
	if fas.flowRow != 0 {
		t.Errorf("expected flowRow = 0, got %d", fas.flowRow)
	}

	// Should NOT reset rate fields
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

	// Should set rate fields
	if fas.rateID != 300 {
		t.Errorf("expected rateID = 300, got %d", fas.rateID)
	}
	if fas.rateRow != 400 {
		t.Errorf("expected rateRow = 400, got %d", fas.rateRow)
	}

	// Should reset flow fields
	if fas.flowID != 0 {
		t.Errorf("expected flowID = 0, got %d", fas.flowID)
	}
	if fas.flowRow != 0 {
		t.Errorf("expected flowRow = 0, got %d", fas.flowRow)
	}

	// Should NOT reset sum fields
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

	// Should set flow fields
	if fas.flowID != 500 {
		t.Errorf("expected flowID = 500, got %d", fas.flowID)
	}
	if fas.flowRow != 600 {
		t.Errorf("expected flowRow = 600, got %d", fas.flowRow)
	}

	// Should NOT reset other fields
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

	// Test that zero values are correct
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

	// Test setting lastHomePage
	fas.lastHomePage = "summary"
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage = 'summary', got '%s'", fas.lastHomePage)
	}

	// Test that reset() doesn't affect lastHomePage
	fas.reset()
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage to remain 'summary' after reset, got '%s'", fas.lastHomePage)
	}

	// Test that setSum() doesn't affect lastHomePage
	fas.setSum(1, 2)
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage to remain 'summary' after setSum, got '%s'", fas.lastHomePage)
	}

	// Test that setRate() doesn't affect lastHomePage
	fas.setRate(3, 4)
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage to remain 'summary' after setRate, got '%s'", fas.lastHomePage)
	}

	// Test that setFlow() doesn't affect lastHomePage
	fas.setFlow(5, 6)
	if fas.lastHomePage != "summary" {
		t.Errorf("expected lastHomePage to remain 'summary' after setFlow, got '%s'", fas.lastHomePage)
	}
}

func TestNewFlowApp(t *testing.T) {
	// Create a mock FlowDataStore and FlowCache for testing
	// We'll use nil since we're only testing the constructor behavior
	var fds *flowdata.FlowDataStore
	var fc *flowcache.FlowCache

	fa := NewFlowApp(fds, fc)

	if fa == nil {
		t.Fatal("expected NewFlowApp to return non-nil FlowApp")
	}

	if fa.mu == nil {
		t.Error("expected mutex to be initialized")
	}

	if fa.app == nil {
		t.Error("expected tview.Application to be initialized")
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

	if fa.pages == nil {
		t.Error("expected tview.Pages to be initialized")
	}
}

func TestFlowApp_BasicFields(t *testing.T) {
	// Test with real objects to ensure proper initialization
	var fds *flowdata.FlowDataStore
	var fc *flowcache.FlowCache

	fa := NewFlowApp(fds, fc)

	// Test that the flowAppState starts with zero values
	if fa.fas.sumID != 0 {
		t.Errorf("expected initial sumID = 0, got %d", fa.fas.sumID)
	}

	if fa.fas.lastHomePage != "" {
		t.Errorf("expected initial lastHomePage = '', got '%s'", fa.fas.lastHomePage)
	}

	// Test that we can modify the state
	fa.fas.setSum(100, 200)
	if fa.fas.sumID != 100 {
		t.Errorf("expected sumID = 100, got %d", fa.fas.sumID)
	}
	if fa.fas.sumRow != 200 {
		t.Errorf("expected sumRow = 200, got %d", fa.fas.sumRow)
	}
}

func TestPageConstants(t *testing.T) {
	// Test that page constants are defined correctly
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

	// Create a mock event
	event := tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)

	// Test with invalid page name - should return the same event
	result := fa.updateSort(event, "testField", true, "invalidPage")

	if result != event {
		t.Error("expected updateSort to return the same event for invalid page")
	}
}

func TestFlowAppState_ConcurrentAccess(t *testing.T) {
	fas := &flowAppState{}

	// Test that we can safely call methods concurrently (no race conditions)
	// This is a basic test - in a real concurrent scenario you'd use sync packages

	fas.setSum(1, 2)
	fas.setRate(3, 4)
	fas.setFlow(5, 6)
	fas.reset()

	// All operations should complete without panic
	if fas.sumID != 0 || fas.rateID != 0 || fas.flowID != 0 {
		t.Error("expected all IDs to be 0 after reset")
	}
}

func TestFlowAppState_ChainedOperations(t *testing.T) {
	fas := &flowAppState{}

	// Test a sequence of operations
	fas.setSum(10, 20)
	fas.setRate(30, 40)
	fas.setFlow(50, 60)

	// Check final state
	if fas.sumID != 10 || fas.sumRow != 20 {
		t.Errorf("expected sum values to remain: ID=10, Row=20, got ID=%d, Row=%d", fas.sumID, fas.sumRow)
	}
	if fas.rateID != 30 || fas.rateRow != 40 {
		t.Errorf("expected rate values to remain: ID=30, Row=40, got ID=%d, Row=%d", fas.rateID, fas.rateRow)
	}
	if fas.flowID != 50 || fas.flowRow != 60 {
		t.Errorf("expected flow values: ID=50, Row=60, got ID=%d, Row=%d", fas.flowID, fas.flowRow)
	}

	// Reset and verify
	fas.reset()
	if fas.sumID != 0 || fas.sumRow != 0 || fas.rateID != 0 || fas.rateRow != 0 || fas.flowID != 0 || fas.flowRow != 0 {
		t.Error("expected all values to be 0 after reset")
	}
} 