package flowcache

import (
	"sync"
	"testing"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
)

// --- Mocks and stubs ---

type mockFlowDataStore struct {
	flowSums     []*flowdata.FlowSum
	flowsBySumID map[int][]*flowdata.FlowData
	calls        map[string]int
	lock         sync.Mutex
}

func (m *mockFlowDataStore) GetFlowSums(_ flowdata.FilterAttributes) []*flowdata.FlowSum {
	m.lock.Lock()
	m.calls["GetFlowSums"]++
	m.lock.Unlock()
	return m.flowSums
}

func (m *mockFlowDataStore) GetFlowsBySumID(sumID int, _ flowdata.FilterAttributes) []*flowdata.FlowData {
	m.lock.Lock()
	m.calls["GetFlowsBySumID"]++
	m.lock.Unlock()
	return m.flowsBySumID[sumID]
}

// --- Test helpers ---

func newMockFlowDataStore() *mockFlowDataStore {
	return &mockFlowDataStore{
		calls:        make(map[string]int),
		flowsBySumID: make(map[int][]*flowdata.FlowData),
	}
}

func setGlobalSort(totalsField string, totalsAsc bool, ratesField string, ratesAsc bool) {
	global.SetSort(flowdata.SortAttributes{
		SumTotalsFieldName: totalsField,
		SumTotalsAscending: totalsAsc,
		SumRatesFieldName:  ratesField,
		SumRatesAscending:  ratesAsc,
	})
}

func setGlobalFilter(filter flowdata.FilterAttributes) {
	global.SetFilter(filter)
}

// --- Tests ---

func TestGetFlowSumTotalsAndRates(t *testing.T) {
	fds := newMockFlowDataStore()
	fds.flowSums = []*flowdata.FlowSum{
		{ID: 1, Key: "a", SourceName: "src1", DestName: "dst1", SourcePacketsIn: 10},
		{ID: 2, Key: "b", SourceName: "src2", DestName: "dst2", SourcePacketsIn: 20},
	}
	setGlobalSort("SourcePacketsIn", true, "ID", false)
	setGlobalFilter(flowdata.FilterAttributes{})

	ctx := t.Context()
	fc := NewFlowCache(ctx, fds)
	// Preload cache
	fc.cacheFlowSums()

	totals := fc.GetFlowSumTotals()
	if len(totals) != 2 || totals[0].SourcePacketsIn > totals[1].SourcePacketsIn {
		t.Errorf("GetFlowSumTotals did not return sorted results: %+v", totals)
	}

	rates := fc.GetFlowSumRates()
	if len(rates) != 2 {
		t.Errorf("GetFlowSumRates did not return expected results: %+v", rates)
	}
}

func TestGetFlowsBySumID(t *testing.T) {
	fds := newMockFlowDataStore()
	fds.flowsBySumID[42] = []*flowdata.FlowData{
		{ID: 1, SumID: 42, FlowResponse: flowdata.FlowResponse{SourceName: "src", DestName: "dst"}},
	}
	setGlobalFilter(flowdata.FilterAttributes{})

	ctx := t.Context()
	fc := NewFlowCache(ctx, fds)

	flows := fc.GetFlowsBySumID(42)
	if len(flows) != 1 || flows[0].SumID != 42 {
		t.Errorf("GetFlowsBySumID did not return expected flows: %+v", flows)
	}
}

func TestCacheRefresh(t *testing.T) {
	fds := newMockFlowDataStore()
	fds.flowSums = []*flowdata.FlowSum{{ID: 1, Key: "a"}}
	setGlobalFilter(flowdata.FilterAttributes{})
	setGlobalSort("", true, "", true)

	ctx := t.Context()
	fc := NewFlowCache(ctx, fds)
	fc.cacheFlowSums()

	// Simulate cache expiry and refresh
	fc.flowSumCache.Remove(flowSumCacheName)
	if got := fc.GetFlowSumTotals(); len(got) != 1 {
		t.Errorf("Cache refresh did not repopulate flow sums: %+v", got)
	}
}

func TestCacheRefreshSorted(t *testing.T) {
	fds := newMockFlowDataStore()
	fds.flowSums = []*flowdata.FlowSum{{ID: 1, Key: "a"}}
	setGlobalFilter(flowdata.FilterAttributes{})
	setGlobalSort("SourcePacketsIn", true, "", true)

	ctx := t.Context()
	fc := NewFlowCache(ctx, fds)
	fc.cacheFlowSums()

	// Simulate cache expiry and refresh
	fc.flowSumCache.Remove(flowSumCacheName)
	if got := fc.GetFlowSumTotals(); len(got) != 1 {
		t.Errorf("Cache refresh sorted did not repopulate flow sums: %+v", got)
	}
}

func TestEmptyCacheReturnsEmptySlices(t *testing.T) {
	fds := newMockFlowDataStore()
	setGlobalFilter(flowdata.FilterAttributes{})
	setGlobalSort("ID", true, "ID", true)

	ctx := t.Context()
	fc := NewFlowCache(ctx, fds)

	if got := fc.GetFlowSumTotals(); len(got) != 0 {
		t.Errorf("Expected empty slice for empty cache, got: %+v", got)
	}
	if got := fc.GetFlowsBySumID(123); len(got) != 0 {
		t.Errorf("Expected empty slice for empty flows, got: %+v", got)
	}
}
