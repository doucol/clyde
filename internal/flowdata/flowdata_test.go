package flowdata

import (
	"testing"
	"time"
)

func TestActionEnum(t *testing.T) {
	// Test Action constants
	expectedActions := map[Action]string{
		Action_ActionUnspecified: "ActionUnspecified",
		Action_Allow:             "Allow",
		Action_Deny:              "Deny",
		Action_Pass:              "Pass",
	}

	for action, expected := range expectedActions {
		if name := Action_name[int32(action)]; name != expected {
			t.Errorf("expected Action_name[%d] = %s, got %s", action, expected, name)
		}
	}

	// Test Action_value map
	expectedValues := map[string]int32{
		"ActionUnspecified": 0,
		"Allow":             1,
		"Deny":              2,
		"Pass":              3,
	}

	for name, expectedValue := range expectedValues {
		if value := Action_value[name]; value != expectedValue {
			t.Errorf("expected Action_value[%s] = %d, got %d", name, expectedValue, value)
		}
	}
}

func TestReporterEnum(t *testing.T) {
	// Test Reporter constants
	expectedReporters := map[Reporter]string{
		Reporter_ReporterUnspecified: "ReporterUnspecified",
		Reporter_Src:                 "Src",
		Reporter_Dst:                 "Dst",
	}

	for reporter, expected := range expectedReporters {
		if name := Reporter_name[int32(reporter)]; name != expected {
			t.Errorf("expected Reporter_name[%d] = %s, got %s", reporter, expected, name)
		}
	}

	// Test Reporter_value map
	expectedValues := map[string]int32{
		"ReporterUnspecified": 0,
		"Src":                 1,
		"Dst":                 2,
	}

	for name, expectedValue := range expectedValues {
		if value := Reporter_value[name]; value != expectedValue {
			t.Errorf("expected Reporter_value[%s] = %d, got %d", name, expectedValue, value)
		}
	}
}

func TestPolicyKindEnum(t *testing.T) {
	// Test a few key PolicyKind constants
	expectedPolicyKinds := map[PolicyKind]string{
		PolicyKind_KindUnspecified:               "KindUnspecified",
		PolicyKind_CalicoNetworkPolicy:           "CalicoNetworkPolicy",
		PolicyKind_GlobalNetworkPolicy:           "GlobalNetworkPolicy",
		PolicyKind_NetworkPolicy:                 "NetworkPolicy",
		PolicyKind_Profile:                       "Profile",
		PolicyKind_EndOfTier:                     "EndOfTier",
	}

	for policyKind, expected := range expectedPolicyKinds {
		if name := PolicyKind_name[int32(policyKind)]; name != expected {
			t.Errorf("expected PolicyKind_name[%d] = %s, got %s", policyKind, expected, name)
		}
	}

	// Test PolicyKind_value map for key entries
	expectedValues := map[string]int32{
		"KindUnspecified":       0,
		"CalicoNetworkPolicy":   1,
		"GlobalNetworkPolicy":   2,
		"NetworkPolicy":         6,
		"Profile":               9,
		"EndOfTier":             10,
	}

	for name, expectedValue := range expectedValues {
		if value := PolicyKind_value[name]; value != expectedValue {
			t.Errorf("expected PolicyKind_value[%s] = %d, got %d", name, expectedValue, value)
		}
	}
}

func TestFlowData_GetID(t *testing.T) {
	fd := &FlowData{ID: 123}
	if fd.GetID() != 123 {
		t.Errorf("expected GetID() = 123, got %d", fd.GetID())
	}
}

func TestFlowData_GetSumKey(t *testing.T) {
	fd := &FlowData{
		FlowResponse: FlowResponse{
			SourceNamespace: "ns1",
			SourceName:      "pod1",
			DestNamespace:   "ns2",
			DestName:        "pod2",
			Protocol:        "TCP",
			DestPort:        80,
		},
	}

	expected := "ns1|pod1|ns2|pod2|TCP|80"
	result := fd.GetSumKey()
	if result != expected {
		t.Errorf("expected GetSumKey() = %s, got %s", expected, result)
	}
}

func TestFlowData_GetterMethods(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(time.Minute)

	fd := &FlowData{
		FlowResponse: FlowResponse{
			StartTime:       startTime,
			EndTime:         endTime,
			Action:          "Allow",
			SourceName:      "test-source",
			SourceNamespace: "test-ns",
			SourceLabels:    "app=test",
			DestName:        "test-dest",
			DestNamespace:   "dest-ns",
			DestLabels:      "app=dest",
			DestPort:        8080,
		},
	}

	tests := []struct {
		name     string
		method   func() interface{}
		expected interface{}
	}{
		{"GetSourceNamespace", func() interface{} { return fd.GetSourceNamespace() }, "test-ns"},
		{"GetSourceName", func() interface{} { return fd.GetSourceName() }, "test-source"},
		{"GetSourceLabels", func() interface{} { return fd.GetSourceLabels() }, "app=test"},
		{"GetDestNamespace", func() interface{} { return fd.GetDestNamespace() }, "dest-ns"},
		{"GetDestName", func() interface{} { return fd.GetDestName() }, "test-dest"},
		{"GetDestLabels", func() interface{} { return fd.GetDestLabels() }, "app=dest"},
		{"GetAction", func() interface{} { return fd.GetAction() }, "Allow"},
		{"GetPort", func() interface{} { return fd.GetPort() }, int64(8080)},
		{"GetStartTime", func() interface{} { return fd.GetStartTime() }, startTime},
		{"GetEndTime", func() interface{} { return fd.GetEndTime() }, endTime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			if result != tt.expected {
				t.Errorf("expected %s() = %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}

func TestFlowSum_GetID(t *testing.T) {
	fs := &FlowSum{ID: 456}
	if fs.GetID() != 456 {
		t.Errorf("expected GetID() = 456, got %d", fs.GetID())
	}
}

func TestFlowSum_GetSumKey(t *testing.T) {
	fs := &FlowSum{Key: "test|sum|key"}
	if fs.GetSumKey() != "test|sum|key" {
		t.Errorf("expected GetSumKey() = 'test|sum|key', got %s", fs.GetSumKey())
	}
}

func TestFlowSum_GetterMethods(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(time.Hour)

	fs := &FlowSum{
		StartTime:       startTime,
		EndTime:         endTime,
		Action:          "Deny",
		SourceName:      "sum-source",
		SourceNamespace: "sum-ns",
		SourceLabels:    "env=prod",
		DestName:        "sum-dest",
		DestNamespace:   "sum-dest-ns",
		DestLabels:      "env=staging",
		DestPort:        443,
	}

	tests := []struct {
		name     string
		method   func() interface{}
		expected interface{}
	}{
		{"GetSourceNamespace", func() interface{} { return fs.GetSourceNamespace() }, "sum-ns"},
		{"GetSourceName", func() interface{} { return fs.GetSourceName() }, "sum-source"},
		{"GetSourceLabels", func() interface{} { return fs.GetSourceLabels() }, "env=prod"},
		{"GetDestNamespace", func() interface{} { return fs.GetDestNamespace() }, "sum-dest-ns"},
		{"GetDestName", func() interface{} { return fs.GetDestName() }, "sum-dest"},
		{"GetDestLabels", func() interface{} { return fs.GetDestLabels() }, "env=staging"},
		{"GetAction", func() interface{} { return fs.GetAction() }, "Deny"},
		{"GetPort", func() interface{} { return fs.GetPort() }, int64(443)},
		{"GetStartTime", func() interface{} { return fs.GetStartTime() }, startTime},
		{"GetEndTime", func() interface{} { return fs.GetEndTime() }, endTime},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.method()
			if result != tt.expected {
				t.Errorf("expected %s() = %v, got %v", tt.name, tt.expected, result)
			}
		})
	}
}

func TestFlowToFlowSum_NewSum(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(time.Minute)

	fd := &FlowData{
		FlowResponse: FlowResponse{
			StartTime:       startTime,
			EndTime:         endTime,
			Action:          "Allow",
			SourceName:      "test-pod",
			SourceNamespace: "test-ns",
			SourceLabels:    "app=test",
			DestName:        "dest-pod",
			DestNamespace:   "dest-ns",
			DestLabels:      "app=dest",
			Protocol:        "TCP",
			DestPort:        80,
			Reporter:        "Src",
			PacketsIn:       100,
			PacketsOut:      50,
			BytesIn:         1000,
			BytesOut:        500,
		},
	}

	result := flowToFlowSum(fd, nil)

	if result == nil {
		t.Fatal("expected flowToFlowSum to return non-nil FlowSum")
	}

	// Verify key properties
	expectedKey := "test-ns|test-pod|dest-ns|dest-pod|TCP|80"
	if result.Key != expectedKey {
		t.Errorf("expected Key = %s, got %s", expectedKey, result.Key)
	}

	if result.Action != "Allow" {
		t.Errorf("expected Action = Allow, got %s", result.Action)
	}

	if result.SourceName != "test-pod" {
		t.Errorf("expected SourceName = test-pod, got %s", result.SourceName)
	}

	// Verify source reporter stats
	if result.SourceReports != 1 {
		t.Errorf("expected SourceReports = 1, got %d", result.SourceReports)
	}

	if result.SourcePacketsIn != 100 {
		t.Errorf("expected SourcePacketsIn = 100, got %d", result.SourcePacketsIn)
	}

	if result.DestReports != 0 {
		t.Errorf("expected DestReports = 0, got %d", result.DestReports)
	}
}

func TestFlowToFlowSum_ExistingSum(t *testing.T) {
	originalStartTime := time.Now().Add(-time.Hour)
	originalEndTime := originalStartTime.Add(30 * time.Minute)
	newStartTime := time.Now().Add(-30 * time.Minute)
	newEndTime := time.Now()

	existing := &FlowSum{
		Key:                 "test|key",
		StartTime:           originalStartTime,
		EndTime:             originalEndTime,
		SourceReports:       1,
		SourcePacketsIn:     50,
		SourceLabels:        "app=test,version=v1",
	}

	fd := &FlowData{
		FlowResponse: FlowResponse{
			StartTime:       newStartTime,
			EndTime:         newEndTime,
			Action:          "Allow",
			SourceName:      "test-pod",
			SourceNamespace: "test-ns",
			SourceLabels:    "app=test,env=prod",
			DestName:        "dest-pod",
			DestNamespace:   "dest-ns",
			Protocol:        "TCP",
			DestPort:        80,
			Reporter:        "Src",
			PacketsIn:       75,
		},
	}

	result := flowToFlowSum(fd, existing)

	if result != existing {
		t.Error("expected flowToFlowSum to return the same FlowSum instance")
	}

	// Verify time ranges are updated correctly
	if !result.StartTime.Equal(originalStartTime) {
		t.Errorf("expected StartTime to be min of original and new, got %v", result.StartTime)
	}

	if !result.EndTime.Equal(newEndTime) {
		t.Errorf("expected EndTime to be max of original and new, got %v", result.EndTime)
	}

	// Verify accumulated stats
	if result.SourceReports != 2 {
		t.Errorf("expected SourceReports = 2, got %d", result.SourceReports)
	}

	if result.SourcePacketsIn != 125 {
		t.Errorf("expected SourcePacketsIn = 125, got %d", result.SourcePacketsIn)
	}
}

func TestFlowToFlowSum_DestReporter(t *testing.T) {
	fd := &FlowData{
		FlowResponse: FlowResponse{
			Reporter:      "Dst",
			PacketsIn:     200,
			PacketsOut:    150,
			BytesIn:       2000,
			BytesOut:      1500,
		},
	}

	result := flowToFlowSum(fd, nil)

	// Verify dest reporter stats
	if result.DestReports != 1 {
		t.Errorf("expected DestReports = 1, got %d", result.DestReports)
	}

	if result.DestPacketsIn != 200 {
		t.Errorf("expected DestPacketsIn = 200, got %d", result.DestPacketsIn)
	}

	if result.DestPacketsOut != 150 {
		t.Errorf("expected DestPacketsOut = 150, got %d", result.DestPacketsOut)
	}

	if result.DestBytesIn != 2000 {
		t.Errorf("expected DestBytesIn = 2000, got %d", result.DestBytesIn)
	}

	if result.DestBytesOut != 1500 {
		t.Errorf("expected DestBytesOut = 1500, got %d", result.DestBytesOut)
	}

	if result.SourceReports != 0 {
		t.Errorf("expected SourceReports = 0, got %d", result.SourceReports)
	}
}

func TestFlowToFlowSum_UnknownReporter(t *testing.T) {
	fd := &FlowData{
		FlowResponse: FlowResponse{
			Reporter: "Unknown",
		},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unknown reporter")
		}
	}()

	flowToFlowSum(fd, nil)
}

func TestFilterAttributes_ZeroValue(t *testing.T) {
	filter := FilterAttributes{}

	// Test zero values
	if filter.Action != "" {
		t.Errorf("expected Action = '', got '%s'", filter.Action)
	}
	if filter.Port != 0 {
		t.Errorf("expected Port = 0, got %d", filter.Port)
	}
	if filter.Namespace != "" {
		t.Errorf("expected Namespace = '', got '%s'", filter.Namespace)
	}
	if filter.Name != "" {
		t.Errorf("expected Name = '', got '%s'", filter.Name)
	}
	if filter.Label != "" {
		t.Errorf("expected Label = '', got '%s'", filter.Label)
	}
	if !filter.DateFrom.IsZero() {
		t.Errorf("expected DateFrom to be zero time, got %v", filter.DateFrom)
	}
	if !filter.DateTo.IsZero() {
		t.Errorf("expected DateTo to be zero time, got %v", filter.DateTo)
	}
}

func TestSortAttributes_ZeroValue(t *testing.T) {
	sort := SortAttributes{}

	// Test zero values
	if sort.SumTotalsFieldName != "" {
		t.Errorf("expected SumTotalsFieldName = '', got '%s'", sort.SumTotalsFieldName)
	}
	if sort.SumTotalsAscending {
		t.Error("expected SumTotalsAscending = false")
	}
	if sort.SumRatesFieldName != "" {
		t.Errorf("expected SumRatesFieldName = '', got '%s'", sort.SumRatesFieldName)
	}
	if sort.SumRatesAscending {
		t.Error("expected SumRatesAscending = false")
	}
}

func TestPolicyTrace_Structure(t *testing.T) {
	// Test that PolicyTrace and PolicyHit structures can be created
	hit := &PolicyHit{
		Kind:        "NetworkPolicy",
		Name:        "test-policy",
		Namespace:   "test-ns",
		Tier:        "default",
		Action:      "Allow",
		PolicyIndex: 1,
		RuleIndex:   2,
	}

	trace := &PolicyTrace{
		Enforced: []*PolicyHit{hit},
		Pending:  []*PolicyHit{},
	}

	if len(trace.Enforced) != 1 {
		t.Errorf("expected 1 enforced policy, got %d", len(trace.Enforced))
	}

	if len(trace.Pending) != 0 {
		t.Errorf("expected 0 pending policies, got %d", len(trace.Pending))
	}

	if trace.Enforced[0].Kind != "NetworkPolicy" {
		t.Errorf("expected Kind = 'NetworkPolicy', got '%s'", trace.Enforced[0].Kind)
	}
}

func TestFlowResponse_Structure(t *testing.T) {
	// Test that FlowResponse can be created with all fields
	startTime := time.Now()
	endTime := startTime.Add(time.Minute)

	fr := &FlowResponse{
		StartTime:       startTime,
		EndTime:         endTime,
		Action:          "Allow",
		SourceName:      "source-pod",
		SourceNamespace: "source-ns",
		SourceLabels:    "app=source",
		DestName:        "dest-pod",
		DestNamespace:   "dest-ns",
		DestLabels:      "app=dest",
		Protocol:        "TCP",
		DestPort:        443,
		Reporter:        "Src",
		Policies:        PolicyTrace{},
		PacketsIn:       100,
		PacketsOut:      50,
		BytesIn:         1000,
		BytesOut:        500,
	}

	if fr.Action != "Allow" {
		t.Errorf("expected Action = 'Allow', got '%s'", fr.Action)
	}

	if fr.DestPort != 443 {
		t.Errorf("expected DestPort = 443, got %d", fr.DestPort)
	}

	if fr.PacketsIn != 100 {
		t.Errorf("expected PacketsIn = 100, got %d", fr.PacketsIn)
	}
} 