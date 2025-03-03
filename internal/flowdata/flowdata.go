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

type FlowAggregate struct {
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
	err = db.Init(&FlowAggregate{})
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

func flowKey(fd *FlowData) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d", fd.SourceNamespace, fd.SourceName, fd.DestNamespace, fd.DestName, fd.Protocol, fd.DestPort)
}

func toAgg(key string, fd *FlowData, fa *FlowAggregate) *FlowAggregate {
	if fa.Key == "" {
		fa.Key = key
		fa.StartTime = fd.StartTime
		fa.EndTime = fd.EndTime
	} else {
		fa.StartTime = util.MinTime(fa.StartTime, fd.StartTime)
		fa.EndTime = util.MaxTime(fa.EndTime, fd.EndTime)
	}
	fa.Action = fd.Action
	fa.SourceName = fd.SourceName
	fa.SourceNamespace = fd.SourceNamespace
	fa.SourceLabels = fd.SourceLabels
	fa.DestName = fd.DestName
	fa.DestNamespace = fd.DestNamespace
	fa.DestLabels = fd.DestLabels
	fa.Protocol = fd.Protocol
	fa.DestPort = fd.DestPort
	fa.Reporter = fd.Reporter
	fa.PacketsIn += uint64(fd.PacketsIn)
	fa.PacketsOut += uint64(fd.PacketsOut)
	fa.BytesIn += uint64(fd.BytesIn)
	fa.BytesOut += uint64(fd.BytesOut)
	return fa
}

func (fds *FlowDataStore) Add(fd *FlowData) error {
	inTx, commit := false, false
	key := flowKey(fd)
	fa := &FlowAggregate{}
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
	err = tx.One("Key", key, fa)
	if err != nil && !errors.Is(err, storm.ErrNotFound) {
		return err
	}
	fa = toAgg(key, fd, fa)
	err = tx.Save(fa)
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

func (fds *FlowDataStore) GetAggregate(id int) (*FlowAggregate, bool) {
	fa := &FlowAggregate{}
	err := fds.db.One("ID", id, fa)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil, false
		} else {
			panic(fmt.Errorf("error getting flow aggregate: %v", err))
		}
	}
	return fa, true
}

func (fds *FlowDataStore) CountAggregate() int {
	cnt, err := fds.db.Count(&FlowAggregate{})
	if err != nil {
		panic(err)
	}
	return cnt
}
