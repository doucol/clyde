package flowdata

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/doucol/clyde/internal/util"
	"k8s.io/apimachinery/pkg/util/runtime"

	storm "github.com/asdine/storm/v3"
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

type FlowSum struct {
	ID               int       `json:"id" storm:"id,increment"`
	Key              string    `json:"key" storm:"unique"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	Action           string    `json:"action"`
	SourceName       string    `json:"source_name"`
	SourceNamespace  string    `json:"source_namespace"`
	SourceLabels     string    `json:"source_labels"`
	DestName         string    `json:"dest_name"`
	DestNamespace    string    `json:"dest_namespace"`
	DestLabels       string    `json:"dest_labels"`
	Protocol         string    `json:"protocol"`
	DestPort         int64     `json:"dest_port"`
	SourceReports    int64     `json:"source_reports"`
	DestReports      int64     `json:"dest_reports"`
	SourcePacketsIn  uint64    `json:"source_packets_in"`
	SourcePacketsOut uint64    `json:"source_packets_out"`
	SourceBytesIn    uint64    `json:"source_bytes_in"`
	SourceBytesOut   uint64    `json:"source_bytes_out"`
	DestPacketsIn    uint64    `json:"dest_packets_in"`
	DestPacketsOut   uint64    `json:"dest_packets_out"`
	DestBytesIn      uint64    `json:"dest_bytes_in"`
	DestBytesOut     uint64    `json:"dest_bytes_out"`
}

type FlowDataStore struct {
	db *storm.DB
}

func dbPath() string {
	return filepath.Join(util.GetDataPath(), "flowdata.db")
}

func NewFlowDataStore() (*FlowDataStore, error) {
	dbPath := dbPath()
	db, err := storm.Open(dbPath)
	if err != nil {
		return nil, err
	}
	err = db.Init(&FlowData{})
	if err != nil {
		return nil, err
	}
	err = db.Init(&FlowSum{})
	if err != nil {
		return nil, err
	}
	return &FlowDataStore{db: db}, nil
}

func Clear() error {
	dbPath := dbPath()
	if util.FileExists(dbPath) {
		return os.Remove(dbPath)
	}
	return nil
}

func (fds *FlowDataStore) Close() {
	runtime.HandleError(fds.db.Close())
}

func (fd *FlowData) Key() string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d", fd.SourceNamespace, fd.SourceName, fd.DestNamespace, fd.DestName, fd.Protocol, fd.DestPort)
}

func toSum(key string, fd *FlowData, fs *FlowSum) *FlowSum {
	if fs == nil {
		fs = &FlowSum{}
		fs.Key = key
		fs.StartTime = fd.StartTime
		fs.EndTime = fd.EndTime
	} else {
		fs.StartTime = util.MinTime(fs.StartTime, fd.StartTime)
		fs.EndTime = util.MaxTime(fs.EndTime, fd.EndTime)
	}
	fs.Action = fd.Action
	fs.SourceName = fd.SourceName
	fs.SourceNamespace = fd.SourceNamespace
	fs.SourceLabels = fd.SourceLabels
	fs.DestName = fd.DestName
	fs.DestNamespace = fd.DestNamespace
	fs.DestLabels = fd.DestLabels
	fs.Protocol = fd.Protocol
	fs.DestPort = fd.DestPort
	switch fd.Reporter {
	case Reporter_name[int32(Reporter_Src)]:
		fs.SourceReports += 1
		fs.SourcePacketsIn += uint64(fd.PacketsIn)
		fs.SourcePacketsOut += uint64(fd.PacketsOut)
		fs.SourceBytesIn += uint64(fd.BytesIn)
		fs.SourceBytesOut += uint64(fd.BytesOut)
	case Reporter_name[int32(Reporter_Dst)]:
		fs.DestReports += 1
		fs.DestPacketsIn += uint64(fd.PacketsIn)
		fs.DestPacketsOut += uint64(fd.PacketsOut)
		fs.DestBytesIn += uint64(fd.BytesIn)
		fs.DestBytesOut += uint64(fd.BytesOut)
	default:
		panic(errors.New("unknown reporter in flow data: " + Reporter_name[int32(Reporter_value[fd.Reporter])]))
	}
	return fs
}

func (fds *FlowDataStore) Add(fd *FlowData) (*FlowSum, bool, error) {
	newSum, inTx, commit := false, false, false
	key := fd.Key()
	fs := &FlowSum{}
	tx, err := fds.db.Begin(true)
	if err != nil {
		return nil, false, err
	}
	inTx = true
	defer func() {
		if inTx && !commit {
			runtime.HandleError(tx.Rollback())
		}
	}()
	err = tx.One("Key", key, fs)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			newSum = true
			fs = nil
		} else {
			return nil, false, err
		}
	}
	fs = toSum(key, fd, fs)
	err = tx.Save(fs)
	if err != nil {
		return nil, false, err
	}
	fd.SumID = fs.ID
	err = tx.Save(fd)
	if err != nil {
		return nil, false, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, false, err
	}
	commit = true
	return fs, newSum, nil
}

func (fds *FlowDataStore) GetFlowSum(id int) (*FlowSum, bool) {
	fs := &FlowSum{}
	err := fds.db.One("ID", id, fs)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil, false
		} else {
			panic(fmt.Errorf("error getting flow aggregate: %v", err))
		}
	}
	return fs, true
}

func (fds *FlowDataStore) GetFlowSumCount() int {
	cnt, err := fds.db.Count(&FlowSum{})
	if err != nil {
		panic(err)
	}
	return cnt
}

func (fds *FlowDataStore) GetFlowDetail(key string, id int) (*FlowData, bool) {
	fs := &FlowSum{}
	err := fds.db.One("Key", key, fs)
	if err != nil {
		panic(fmt.Errorf("error getting flow aggregate: %v", err))
	}

	fd := []FlowData{}
	err = fds.db.Find("SumID", fs.ID, &fd, storm.Skip(id-1), storm.Limit(1))
	if err != nil {
		panic(fmt.Errorf("error getting flow detail: %s, %d, %v", key, id, fs))
	}
	return &fd[0], true
}

func (fds *FlowDataStore) GetFlowDetailCount(key string) int {
	// TODO:: this is horrible - need to rewrite this soon
	fs := &FlowSum{}
	err := fds.db.One("Key", key, fs)
	if err != nil {
		panic(fmt.Errorf("error getting flow aggregate: %v", err))
	}

	fd := []FlowData{}
	err = fds.db.Find("SumID", fs.ID, &fd)
	if err != nil {
		panic(fmt.Errorf("error getting flow detail: %s, %v", key, fs))
	}
	return len(fd)
}
