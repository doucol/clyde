package flowdata

import (
	"fmt"
	"strings"
	"time"
)

type Action int32

const (
	Action_ActionUnspecified Action = 0
	Action_Allow             Action = 1
	Action_Deny              Action = 2
	Action_Pass              Action = 3
)

// Enum value maps for Action.
var (
	Action_name = map[int32]string{
		0: "ActionUnspecified",
		1: "Allow",
		2: "Deny",
		3: "Pass",
	}
	Action_value = map[string]int32{
		"ActionUnspecified": 0,
		"Allow":             1,
		"Deny":              2,
		"Pass":              3,
	}
)

type Reporter int32

const (
	// For queries, unspecified means "do not filter on this field".
	Reporter_ReporterUnspecified Reporter = 0
	Reporter_Src                 Reporter = 1
	Reporter_Dst                 Reporter = 2
)

// Enum value maps for Reporter.
var (
	Reporter_name = map[int32]string{
		0: "ReporterUnspecified",
		1: "Src",
		2: "Dst",
	}
	Reporter_value = map[string]int32{
		"ReporterUnspecified": 0,
		"Src":                 1,
		"Dst":                 2,
	}
)

type PolicyKind int32

const (
	// Unspecified
	PolicyKind_KindUnspecified PolicyKind = 0
	// Calico policy types.
	PolicyKind_CalicoNetworkPolicy           PolicyKind = 1
	PolicyKind_GlobalNetworkPolicy           PolicyKind = 2
	PolicyKind_StagedNetworkPolicy           PolicyKind = 3
	PolicyKind_StagedGlobalNetworkPolicy     PolicyKind = 4
	PolicyKind_StagedKubernetesNetworkPolicy PolicyKind = 5
	// Native Kubernetes types.
	PolicyKind_NetworkPolicy              PolicyKind = 6
	PolicyKind_AdminNetworkPolicy         PolicyKind = 7
	PolicyKind_BaselineAdminNetworkPolicy PolicyKind = 8
	// Calico Profiles.
	PolicyKind_Profile   PolicyKind = 9
	PolicyKind_EndOfTier PolicyKind = 10
)

// Enum value maps for PolicyKind.
var (
	PolicyKind_name = map[int32]string{
		0:  "KindUnspecified",
		1:  "CalicoNetworkPolicy",
		2:  "GlobalNetworkPolicy",
		3:  "StagedNetworkPolicy",
		4:  "StagedGlobalNetworkPolicy",
		5:  "StagedKubernetesNetworkPolicy",
		6:  "NetworkPolicy",
		7:  "AdminNetworkPolicy",
		8:  "BaselineAdminNetworkPolicy",
		9:  "Profile",
		10: "EndOfTier",
	}
	PolicyKind_value = map[string]int32{
		"KindUnspecified":               0,
		"CalicoNetworkPolicy":           1,
		"GlobalNetworkPolicy":           2,
		"StagedNetworkPolicy":           3,
		"StagedGlobalNetworkPolicy":     4,
		"StagedKubernetesNetworkPolicy": 5,
		"NetworkPolicy":                 6,
		"AdminNetworkPolicy":            7,
		"BaselineAdminNetworkPolicy":    8,
		"Profile":                       9,
		"EndOfTier":                     10,
	}
)

type PolicyTrace struct {
	Enforced []*PolicyHit `json:"enforced"`
	Pending  []*PolicyHit `json:"pending"`
}

type PolicyHit struct {
	Kind        string     `json:"kind"`
	Name        string     `json:"name"`
	Namespace   string     `json:"namespace"`
	Tier        string     `json:"tier"`
	Action      string     `json:"action"`
	PolicyIndex int64      `json:"policy_index"`
	RuleIndex   int64      `json:"rule_index"`
	Trigger     *PolicyHit `json:"trigger"`
}

type FlowResponse struct {
	StartTime       time.Time   `json:"start_time"`
	EndTime         time.Time   `json:"end_time"`
	Action          string      `json:"action"`
	SourceName      string      `json:"source_name"`
	SourceNamespace string      `json:"source_namespace"`
	SourceLabels    string      `json:"source_labels"`
	DestName        string      `json:"dest_name"`
	DestNamespace   string      `json:"dest_namespace"`
	DestLabels      string      `json:"dest_labels"`
	Protocol        string      `json:"protocol"`
	DestPort        int64       `json:"dest_port"`
	Reporter        string      `json:"reporter"`
	Policies        PolicyTrace `json:"policies"`
	PacketsIn       int64       `json:"packets_in"`
	PacketsOut      int64       `json:"packets_out"`
	BytesIn         int64       `json:"bytes_in"`
	BytesOut        int64       `json:"bytes_out"`
}

type FlowData struct {
	ID           int `json:"id" storm:"id,increment"`
	SumID        int `json:"sum_id" storm:"index"`
	FlowResponse `storm:"inline"`
}

type FilterAttributes struct {
	Action    string
	Port      int
	Namespace string
	Name      string
	Label     string
	DateFrom  time.Time
	DateTo    time.Time
}

type SortAttributes struct {
	SumTotalsFieldName string
	SumTotalsAscending bool
	SumRatesFieldName  string
	SumRatesAscending  bool
}

// [Flower] interface
func (fd *FlowData) GetID() int {
	return fd.ID
}

func (fd *FlowData) GetSumKey() string {
	key := []string{fd.SourceNamespace, fd.SourceName, fd.DestNamespace, fd.DestName, fd.Protocol, fmt.Sprint(fd.DestPort)}
	return strings.Join(key, "|")
}

func (fd *FlowData) GetSourceNamespace() string {
	return fd.SourceNamespace
}

func (fd *FlowData) GetSourceName() string {
	return fd.SourceName
}

func (fd *FlowData) GetSourceLabels() string {
	return fd.SourceLabels
}

func (fd *FlowData) GetDestNamespace() string {
	return fd.DestNamespace
}

func (fd *FlowData) GetDestName() string {
	return fd.DestName
}

func (fd *FlowData) GetDestLabels() string {
	return fd.DestLabels
}

func (fd *FlowData) GetAction() string {
	return fd.Action
}

func (fd *FlowData) GetPort() int64 {
	return fd.DestPort
}

func (fd *FlowData) GetStartTime() time.Time {
	return fd.StartTime
}

func (fd *FlowData) GetEndTime() time.Time {
	return fd.EndTime
}
