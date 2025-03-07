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

type FlowResponse struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Action          string    `json:"action"`
	SourceName      string    `json:"source_name"`
	SourceNamespace string    `json:"source_namespace"`
	SourceLabels    string    `json:"source_labels"`
	DestName        string    `json:"dest_name"`
	DestNamespace   string    `json:"dest_namespace"`
	DestLabels      string    `json:"dest_labels"`
	Protocol        string    `json:"protocol"`
	DestPort        int64     `json:"dest_port"`
	Reporter        string    `json:"reporter"`
	PacketsIn       int64     `json:"packets_in"`
	PacketsOut      int64     `json:"packets_out"`
	BytesIn         int64     `json:"bytes_in"`
	BytesOut        int64     `json:"bytes_out"`
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

func flowSumKey(fd *FlowData) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d", fd.SourceNamespace, fd.SourceName, fd.DestNamespace, fd.DestName, fd.Protocol, fd.DestPort)
}

func toSum(key string, fd *FlowData, fs *FlowSum) *FlowSum {
	if fs.Key == "" {
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
	case "src":
		fs.SourceReports += 1
		fs.SourcePacketsIn += uint64(fd.PacketsIn)
		fs.SourcePacketsOut += uint64(fd.PacketsOut)
		fs.SourceBytesIn += uint64(fd.BytesIn)
		fs.SourceBytesOut += uint64(fd.BytesOut)
	case "dst":
		fs.DestReports += 1
		fs.DestPacketsIn += uint64(fd.PacketsIn)
		fs.DestPacketsOut += uint64(fd.PacketsOut)
		fs.DestBytesIn += uint64(fd.BytesIn)
		fs.DestBytesOut += uint64(fd.BytesOut)
	default:
		panic(errors.New("unknown reporter in flow data: " + fd.Reporter))
	}
	return fs
}

func (fds *FlowDataStore) Add(fd *FlowData) error {
	inTx, commit := false, false
	key := flowSumKey(fd)
	fs := &FlowSum{}
	tx, err := fds.db.Begin(true)
	if err != nil {
		return err
	}
	inTx = true
	defer func() {
		if inTx && !commit {
			runtime.HandleError(tx.Rollback())
		}
	}()
	err = tx.One("Key", key, fs)
	if err != nil && !errors.Is(err, storm.ErrNotFound) {
		return err
	}
	fs = toSum(key, fd, fs)
	err = tx.Save(fs)
	if err != nil {
		return err
	}
	fd.SumID = fs.ID
	err = tx.Save(fd)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	commit = true
	return nil
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
