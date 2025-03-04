package flowdata

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/doucol/clyde/internal/util"

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
	FlowResponse `storm:"inline"`
}

type FlowSum struct {
	ID              int       `json:"id" storm:"id,increment"`
	Key             string    `json:"key" storm:"unique"`
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
	PacketsIn       uint64    `json:"packets_in"`
	PacketsOut      uint64    `json:"packets_out"`
	BytesIn         uint64    `json:"bytes_in"`
	BytesOut        uint64    `json:"bytes_out"`
}

type FlowDataStore struct {
	db *storm.DB
}

func NewFlowDataStore() (*FlowDataStore, error) {
	dbPath := filepath.Join(util.GetDataPath(), "flowdata.db")
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

func (fds *FlowDataStore) Close() {
	err := fds.db.Close()
	if err != nil {
		panic(err)
	}
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
	fs.Reporter = fd.Reporter
	fs.PacketsIn += uint64(fd.PacketsIn)
	fs.PacketsOut += uint64(fd.PacketsOut)
	fs.BytesIn += uint64(fd.BytesIn)
	fs.BytesOut += uint64(fd.BytesOut)
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
			err := tx.Rollback()
			if err != nil {
				panic(fmt.Sprintf("error rolling back transaction: %v", err))
			}
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

func (fds *FlowDataStore) GetFlowDetail(id int) (*FlowData, bool) {
	fd := &FlowData{}
	err := fds.db.One("ID", id, fd)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil, false
		} else {
			panic(fmt.Errorf("error getting flow aggregate: %v", err))
		}
	}
	return fd, true
}

func (fds *FlowDataStore) GetFlowDetailCount(key string) int {
	cnt, err := fds.db.Count(&FlowData{})
	if err != nil {
		panic(err)
	}
	return cnt
}
