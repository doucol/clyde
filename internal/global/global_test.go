package global

import (
	"testing"

	"github.com/doucol/clyde/internal/flowdata"
)

func TestGlobalState(t *testing.T) {
	// Test initial state
	initialState := GetState()
	if initialState.Filter != (flowdata.FilterAttributes{}) {
		t.Errorf("Initial filter state is not empty: %v", initialState.Filter)
	}
	if initialState.Sort != (flowdata.SortAttributes{}) {
		t.Errorf("Initial sort state is not empty: %v", initialState.Sort)
	}

	// Test setting and getting state
	newState := GlobalState{
		Filter: flowdata.FilterAttributes{
			Namespace: "test-namespace",
			Name:      "test-name",
		},
		Sort: flowdata.SortAttributes{
			SumTotalsFieldName: "test-field",
			SumTotalsAscending: true,
			SumRatesFieldName:  "test-rate-field",
			SumRatesAscending:  false,
		},
	}

	SetState(newState)
	gotState := GetState()

	if gotState.Filter.Namespace != newState.Filter.Namespace {
		t.Errorf("Filter.Namespace = %v; want %v", gotState.Filter.Namespace, newState.Filter.Namespace)
	}
	if gotState.Filter.Name != newState.Filter.Name {
		t.Errorf("Filter.Name = %v; want %v", gotState.Filter.Name, newState.Filter.Name)
	}
	if gotState.Sort.SumTotalsFieldName != newState.Sort.SumTotalsFieldName {
		t.Errorf("Sort.SumTotalsFieldName = %v; want %v", gotState.Sort.SumTotalsFieldName, newState.Sort.SumTotalsFieldName)
	}
	if gotState.Sort.SumTotalsAscending != newState.Sort.SumTotalsAscending {
		t.Errorf("Sort.SumTotalsAscending = %v; want %v", gotState.Sort.SumTotalsAscending, newState.Sort.SumTotalsAscending)
	}
	if gotState.Sort.SumRatesFieldName != newState.Sort.SumRatesFieldName {
		t.Errorf("Sort.SumRatesFieldName = %v; want %v", gotState.Sort.SumRatesFieldName, newState.Sort.SumRatesFieldName)
	}
	if gotState.Sort.SumRatesAscending != newState.Sort.SumRatesAscending {
		t.Errorf("Sort.SumRatesAscending = %v; want %v", gotState.Sort.SumRatesAscending, newState.Sort.SumRatesAscending)
	}
}

func TestFilterAndSort(t *testing.T) {
	// Test filter operations
	filter := flowdata.FilterAttributes{
		Namespace: "test-namespace",
		Name:      "test-name",
	}

	SetFilter(filter)
	gotFilter := GetFilter()

	if gotFilter.Namespace != filter.Namespace {
		t.Errorf("GetFilter().Namespace = %v; want %v", gotFilter.Namespace, filter.Namespace)
	}
	if gotFilter.Name != filter.Name {
		t.Errorf("GetFilter().Name = %v; want %v", gotFilter.Name, filter.Name)
	}

	// Test sort operations
	sort := flowdata.SortAttributes{
		SumTotalsFieldName: "test-field",
		SumTotalsAscending: true,
	}

	SetSort(sort)
	gotSort := GetSort()

	if gotSort.SumTotalsFieldName != sort.SumTotalsFieldName {
		t.Errorf("GetSort().SumTotalsFieldName = %v; want %v", gotSort.SumTotalsFieldName, sort.SumTotalsFieldName)
	}
	if gotSort.SumTotalsAscending != sort.SumTotalsAscending {
		t.Errorf("GetSort().SumTotalsAscending = %v; want %v", gotSort.SumTotalsAscending, sort.SumTotalsAscending)
	}
}
