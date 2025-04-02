package flowdata

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/doucol/clyde/internal/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type FlowDataStore struct {
	db *storm.DB
}

type Flower interface {
	GetID() int
	GetSumKey() string
	GetSourceNamespace() string
	GetSourceName() string
	GetSourceLabels() string
	GetDestNamespace() string
	GetDestName() string
	GetDestLabels() string
	GetPort() int64
	GetAction() string
	GetStartTime() time.Time
	GetEndTime() time.Time
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

func (fds *FlowDataStore) AddFlow(fd *FlowData) (*FlowSum, bool, error) {
	newSum, committed := false, false
	fs := &FlowSum{}
	tx, err := fds.db.Begin(true)
	if err != nil {
		return nil, false, err
	}
	defer func() {
		if !committed {
			runtime.HandleError(tx.Rollback())
		}
	}()
	err = tx.One("Key", fd.GetSumKey(), fs)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			newSum = true
			fs = nil
		} else {
			return nil, false, err
		}
	}
	fs = flowToFlowSum(fd, fs)
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
	committed = true
	return fs, newSum, nil
}

// func (fds *FlowDataStore) aggregateMetrics(tx storm.Node, fd *FlowData, fs *FlowSum) {
// 	var srcPacketsInSum, srcPacketsOutSum, srcBytesInSum, srcBytesOutSum uint64
// 	var destPacketsInSum, destPacketsOutSum, destBytesInSum, destBytesOutSum uint64
// 	agg := func(f *FlowData) {
// 		reporter := strings.ToLower(f.Reporter)
// 		switch reporter {
// 		case "src":
// 			srcPacketsInSum += uint64(f.PacketsIn)
// 			srcPacketsOutSum += uint64(f.PacketsOut)
// 			srcBytesInSum += uint64(f.BytesIn)
// 			srcBytesOutSum += uint64(f.BytesOut)
// 		case "dst":
// 			destPacketsInSum += uint64(f.PacketsIn)
// 			destPacketsOutSum += uint64(f.PacketsOut)
// 			destBytesInSum += uint64(f.BytesIn)
// 			destBytesOutSum += uint64(f.BytesOut)
// 		}
// 	}
// 	fa := FilterAttributes{DateFrom: time.Now().UTC().Add(-5 * time.Minute)}
// 	for _, sample := range fds.GetFlowsBySumID(fs.ID, fa, tx) {
// 		agg(sample)
// 	}
// 	agg(fd)
// }

func (fds *FlowDataStore) GetFlowSum(id int) *FlowSum {
	fs := &FlowSum{}
	err := fds.db.One("ID", id, fs)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil
		} else {
			logrus.WithError(err).Panic("error getting flow sum")
		}
	}
	return fs
}

func (fds *FlowDataStore) GetFlowSums(filter FilterAttributes) []*FlowSum {
	fs := []*FlowSum{}
	err := fds.db.AllByIndex("Key", &fs)
	if err != nil && !errors.Is(err, storm.ErrNotFound) {
		logrus.WithError(err).Panic("error getting all flow sums")
	}
	if filter != (FilterAttributes{}) {
		fs = util.FilterSlice(fs, func(f *FlowSum) bool {
			return filterFlow(f, filter)
		})
	}
	return fs
}

func (fds *FlowDataStore) GetFlowDetail(id int) *FlowData {
	fd := &FlowData{}
	err := fds.db.One("ID", id, fd)
	if err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			return nil
		} else {
			logrus.WithError(err).Panic("error getting flow data")
		}
	}
	return fd
}

func (fds *FlowDataStore) GetFlowsBySumID(sumID int, filter FilterAttributes) []*FlowData {
	fd := []*FlowData{}
	err := fds.db.Find("SumID", sumID, &fd)
	if err != nil && !errors.Is(err, storm.ErrNotFound) {
		logrus.WithError(err).Panic("error getting all flow sums")
	}
	if filter != (FilterAttributes{}) {
		fd = util.FilterSlice(fd, func(f *FlowData) bool {
			return filterFlow(f, filter)
		})
	}
	return fd
}

func filterFlow(f Flower, filter FilterAttributes) bool {
	// These checks are just for FlowSum
	if _, ok := f.(*FlowSum); ok {
		if filter.Port > 0 && f.GetPort() != int64(filter.Port) {
			return false
		}
		if filter.Namespace != "" {
			if !strings.Contains(f.GetSourceNamespace(), filter.Namespace) && !strings.Contains(f.GetDestNamespace(), filter.Namespace) {
				return false
			}
		}
		if filter.Name != "" {
			if !strings.Contains(f.GetSourceName(), filter.Name) && !strings.Contains(f.GetDestName(), filter.Name) {
				return false
			}
		}
	}
	// These checks are for FlowSum and FlowData
	if filter.Action != "" && f.GetAction() != filter.Action {
		return false
	}
	if filter.Label != "" {
		if !strings.Contains(f.GetSourceLabels(), filter.Label) && !strings.Contains(f.GetDestLabels(), filter.Label) {
			return false
		}
	}
	if !filter.DateFrom.IsZero() && !filter.DateTo.IsZero() {
		if f.GetEndTime().Before(filter.DateFrom) || f.GetStartTime().After(filter.DateTo) {
			return false
		}
	} else if !filter.DateFrom.IsZero() && f.GetEndTime().Before(filter.DateFrom) {
		return false
	} else if !filter.DateTo.IsZero() && f.GetStartTime().After(filter.DateTo) {
		return false
	}
	return true
}
