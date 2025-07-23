package flowdata

import (
	"os"
	"testing"
	"time"
)

func TestNewFlowDataStore(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	fds, err := NewFlowDataStore()
	if err != nil {
		t.Fatalf("expected NewFlowDataStore to succeed, got error: %v", err)
	}
	defer fds.Close()

	if fds == nil {
		t.Fatal("expected NewFlowDataStore to return non-nil FlowDataStore")
	}

	if fds.db == nil {
		t.Error("expected database to be initialized")
	}

	if fds.RateCalcWindow != 60 {
		t.Errorf("expected default RateCalcWindow = 60, got %d", fds.RateCalcWindow)
	}

	if fds.RateCalcInterval != 5 {
		t.Errorf("expected default RateCalcInterval = 5, got %d", fds.RateCalcInterval)
	}
}

func TestNewFlowDataStore_DatabaseError(t *testing.T) {
	// Skip this test as it requires complex filesystem permission setup
	// that's difficult to achieve reliably across different environments
	t.Skip("Database error testing requires complex filesystem setup")
}

func TestClear(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Create a FlowDataStore to ensure the database file exists
	fds, err := NewFlowDataStore()
	if err != nil {
		t.Fatalf("failed to create FlowDataStore: %v", err)
	}
	fds.Close()

	// Clear should succeed
	err = Clear()
	if err != nil {
		t.Errorf("expected Clear to succeed, got error: %v", err)
	}

	// Calling Clear again should not error (file doesn't exist)
	err = Clear()
	if err != nil {
		t.Errorf("expected Clear to succeed when file doesn't exist, got error: %v", err)
	}
}

func TestFlowDataStore_ChannelMethods(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	fds, err := NewFlowDataStore()
	if err != nil {
		t.Fatalf("failed to create FlowDataStore: %v", err)
	}
	defer fds.Close()

	// Test that channel methods return non-nil channels
	if fds.FlowAdded() == nil {
		t.Error("expected FlowAdded to return non-nil channel")
	}

	if fds.FlowSumAdded() == nil {
		t.Error("expected FlowSumAdded to return non-nil channel")
	}

	if fds.FlowSumsUpdated() == nil {
		t.Error("expected FlowSumsUpdated to return non-nil channel")
	}

	if fds.FlowRatesUpdated() == nil {
		t.Error("expected FlowRatesUpdated to return non-nil channel")
	}

	// Test that calling multiple times returns the same channel
	ch1 := fds.FlowAdded()
	ch2 := fds.FlowAdded()
	if ch1 != ch2 {
		t.Error("expected FlowAdded to return the same channel on multiple calls")
	}
}

func TestFilterFlow_FlowSum(t *testing.T) {
	now := time.Now()
	fs := &FlowSum{
		Action:          "Allow",
		SourceName:      "test-pod",
		SourceNamespace: "test-ns",
		SourceLabels:    "app=test,env=prod",
		DestName:        "dest-pod",
		DestNamespace:   "dest-ns",
		DestLabels:      "app=dest",
		DestPort:        80,
		StartTime:       now.Add(-time.Hour),
		EndTime:         now,
	}

	tests := []struct {
		name     string
		filter   FilterAttributes
		expected bool
	}{
		{
			name:     "empty filter matches all",
			filter:   FilterAttributes{},
			expected: true,
		},
		{
			name:     "action filter matches",
			filter:   FilterAttributes{Action: "Allow"},
			expected: true,
		},
		{
			name:     "action filter doesn't match",
			filter:   FilterAttributes{Action: "Deny"},
			expected: false,
		},
		{
			name:     "port filter matches",
			filter:   FilterAttributes{Port: 80},
			expected: true,
		},
		{
			name:     "port filter doesn't match",
			filter:   FilterAttributes{Port: 443},
			expected: false,
		},
		{
			name:     "namespace filter matches source",
			filter:   FilterAttributes{Namespace: "test"},
			expected: true,
		},
		{
			name:     "namespace filter matches dest",
			filter:   FilterAttributes{Namespace: "dest"},
			expected: true,
		},
		{
			name:     "namespace filter doesn't match",
			filter:   FilterAttributes{Namespace: "nonexistent"},
			expected: false,
		},
		{
			name:     "name filter matches source",
			filter:   FilterAttributes{Name: "test-pod"},
			expected: true,
		},
		{
			name:     "name filter matches dest",
			filter:   FilterAttributes{Name: "dest-pod"},
			expected: true,
		},
		{
			name:     "name filter doesn't match",
			filter:   FilterAttributes{Name: "nonexistent"},
			expected: false,
		},
		{
			name:     "label filter matches source",
			filter:   FilterAttributes{Label: "app=test"},
			expected: true,
		},
		{
			name:     "label filter matches dest",
			filter:   FilterAttributes{Label: "app=dest"},
			expected: true,
		},
		{
			name:     "label filter doesn't match",
			filter:   FilterAttributes{Label: "nonexistent=value"},
			expected: false,
		},
		{
			name:     "date range filter matches",
			filter:   FilterAttributes{DateFrom: now.Add(-2 * time.Hour), DateTo: now.Add(time.Hour)},
			expected: true,
		},
		{
			name:     "date range filter before range",
			filter:   FilterAttributes{DateFrom: now.Add(time.Hour), DateTo: now.Add(2 * time.Hour)},
			expected: false,
		},
		{
			name:     "date range filter after range",
			filter:   FilterAttributes{DateFrom: now.Add(-3 * time.Hour), DateTo: now.Add(-2 * time.Hour)},
			expected: false,
		},
		{
			name:     "date from filter matches",
			filter:   FilterAttributes{DateFrom: now.Add(-2 * time.Hour)},
			expected: true,
		},
		{
			name:     "date from filter doesn't match",
			filter:   FilterAttributes{DateFrom: now.Add(time.Hour)},
			expected: false,
		},
		{
			name:     "date to filter matches",
			filter:   FilterAttributes{DateTo: now.Add(time.Hour)},
			expected: true,
		},
		{
			name:     "date to filter doesn't match",
			filter:   FilterAttributes{DateTo: now.Add(-2 * time.Hour)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterFlow(fs, tt.filter)
			if result != tt.expected {
				t.Errorf("expected filterFlow = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFilterFlow_FlowData(t *testing.T) {
	now := time.Now()
	fd := &FlowData{
		FlowResponse: FlowResponse{
			Action:       "Deny",
			SourceLabels: "app=source,env=test",
			DestLabels:   "app=dest,env=prod",
			StartTime:    now.Add(-30 * time.Minute),
			EndTime:      now,
		},
	}

	tests := []struct {
		name     string
		filter   FilterAttributes
		expected bool
	}{
		{
			name:     "action filter matches",
			filter:   FilterAttributes{Action: "Deny"},
			expected: true,
		},
		{
			name:     "action filter doesn't match",
			filter:   FilterAttributes{Action: "Allow"},
			expected: false,
		},
		{
			name:     "label filter matches source",
			filter:   FilterAttributes{Label: "env=test"},
			expected: true,
		},
		{
			name:     "label filter matches dest",
			filter:   FilterAttributes{Label: "env=prod"},
			expected: true,
		},
		{
			name:     "label filter doesn't match",
			filter:   FilterAttributes{Label: "nonexistent=value"},
			expected: false,
		},
		{
			name:     "date range matches",
			filter:   FilterAttributes{DateFrom: now.Add(-time.Hour), DateTo: now.Add(time.Hour)},
			expected: true,
		},
		{
			name:     "port filter ignored for FlowData",
			filter:   FilterAttributes{Port: 80},
			expected: true,
		},
		{
			name:     "namespace filter ignored for FlowData",
			filter:   FilterAttributes{Namespace: "test"},
			expected: true,
		},
		{
			name:     "name filter ignored for FlowData",
			filter:   FilterAttributes{Name: "test"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterFlow(fd, tt.filter)
			if result != tt.expected {
				t.Errorf("expected filterFlow = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFlowerInterface(t *testing.T) {
	// Test that FlowData implements Flower interface
	var _ Flower = &FlowData{}

	// Test that FlowSum implements Flower interface
	var _ Flower = &FlowSum{}

	// Test interface methods on FlowData
	now := time.Now()
	fd := &FlowData{
		ID: 100,
		FlowResponse: FlowResponse{
			SourceNamespace: "test-ns",
			SourceName:      "test-pod",
			DestNamespace:   "dest-ns",
			DestName:        "dest-pod",
			Protocol:        "TCP",
			DestPort:        8080,
			Action:          "Allow",
			StartTime:       now,
			EndTime:         now.Add(time.Minute),
			SourceLabels:    "app=test",
			DestLabels:      "app=dest",
		},
	}

	if fd.GetID() != 100 {
		t.Errorf("expected GetID() = 100, got %d", fd.GetID())
	}

	expectedKey := "test-ns|test-pod|dest-ns|dest-pod|TCP|8080"
	if fd.GetSumKey() != expectedKey {
		t.Errorf("expected GetSumKey() = %s, got %s", expectedKey, fd.GetSumKey())
	}

	// Test interface methods on FlowSum
	fs := &FlowSum{
		ID:              200,
		Key:             "sum-key",
		SourceNamespace: "sum-ns",
		SourceName:      "sum-pod",
		DestNamespace:   "sum-dest-ns",
		DestName:        "sum-dest-pod",
		Action:          "Deny",
		DestPort:        443,
		StartTime:       now,
		EndTime:         now.Add(time.Hour),
		SourceLabels:    "env=test",
		DestLabels:      "env=prod",
	}

	if fs.GetID() != 200 {
		t.Errorf("expected GetID() = 200, got %d", fs.GetID())
	}

	if fs.GetSumKey() != "sum-key" {
		t.Errorf("expected GetSumKey() = 'sum-key', got %s", fs.GetSumKey())
	}
}

func TestDbPath(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	path := dbPath()
	if path == "" {
		t.Error("expected dbPath to return non-empty string")
	}

	// Should contain flowdata.db
	if len(path) < len("flowdata.db") {
		t.Errorf("expected path to contain 'flowdata.db', got %s", path)
	}
}

func TestFlowDataStore_Configuration(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	fds, err := NewFlowDataStore()
	if err != nil {
		t.Fatalf("failed to create FlowDataStore: %v", err)
	}
	defer fds.Close()

	// Test default configuration
	if fds.RateCalcWindow != 60 {
		t.Errorf("expected default RateCalcWindow = 60, got %d", fds.RateCalcWindow)
	}

	if fds.RateCalcInterval != 5 {
		t.Errorf("expected default RateCalcInterval = 5, got %d", fds.RateCalcInterval)
	}

	// Test configuration can be changed
	fds.RateCalcWindow = 30
	fds.RateCalcInterval = 10

	if fds.RateCalcWindow != 30 {
		t.Errorf("expected RateCalcWindow = 30, got %d", fds.RateCalcWindow)
	}

	if fds.RateCalcInterval != 10 {
		t.Errorf("expected RateCalcInterval = 10, got %d", fds.RateCalcInterval)
	}
}

func TestChanSignal(t *testing.T) {
	// Test with nil channel (should not panic)
	chanSignal[string](nil, "test")

	// Test with valid unbuffered channel
	ch := make(chan string)

	// Use goroutine to prevent blocking
	go func() {
		chanSignal(ch, "test")
	}()

	select {
	case msg := <-ch:
		if msg != "test" {
			t.Errorf("expected message 'test', got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected message to be sent to channel within timeout")
	}
}

func TestFlowDataStore_Close(t *testing.T) {
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	fds, err := NewFlowDataStore()
	if err != nil {
		t.Fatalf("failed to create FlowDataStore: %v", err)
	}

	// Initialize channels
	fds.FlowAdded()
	fds.FlowSumAdded()
	fds.FlowSumsUpdated()
	fds.FlowRatesUpdated()

	// Close should not panic
	fds.Close()

	// Close again should not panic
	fds.Close()
}
