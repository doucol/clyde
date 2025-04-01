package flowdata

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/asdine/storm/v3"
	"github.com/doucol/clyde/internal/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type FlowDataStore struct {
	db *storm.DB
}

type FlowItem interface {
	GetID() int
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
	err = tx.One("Key", fd.SumKey(), fs)
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
			if filter.Action != "" && f.Action != filter.Action {
				return false
			}
			if filter.Port > 0 && f.DestPort != int64(filter.Port) {
				return false
			}
			if filter.Namespace != "" {
				if !strings.Contains(f.SourceNamespace, filter.Namespace) && !strings.Contains(f.DestNamespace, filter.Namespace) {
					return false
				}
			}
			if filter.Name != "" {
				if !strings.Contains(f.SourceName, filter.Name) && !strings.Contains(f.DestName, filter.Name) {
					return false
				}
			}
			if filter.Label != "" {
				if !strings.Contains(f.SourceLabels, filter.Label) && !strings.Contains(f.DestLabels, filter.Label) {
					return false
				}
			}
			if !filter.DateFrom.IsZero() && !filter.DateTo.IsZero() {
				if f.EndTime.Before(filter.DateFrom) || f.StartTime.After(filter.DateTo) {
					return false
				}
			} else if !filter.DateFrom.IsZero() && f.EndTime.Before(filter.DateFrom) {
				return false
			} else if !filter.DateTo.IsZero() && f.StartTime.After(filter.DateTo) {
				return false
			}
			return true
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
			if filter.Action != "" && f.Action != filter.Action {
				return false
			}
			if filter.Label != "" {
				if !strings.Contains(f.SourceLabels, filter.Label) && !strings.Contains(f.DestLabels, filter.Label) {
					return false
				}
			}
			if !filter.DateFrom.IsZero() && !filter.DateTo.IsZero() {
				if f.EndTime.Before(filter.DateFrom) || f.StartTime.After(filter.DateTo) {
					return false
				}
			} else if !filter.DateFrom.IsZero() && f.EndTime.Before(filter.DateFrom) {
				return false
			} else if !filter.DateTo.IsZero() && f.StartTime.After(filter.DateTo) {
				return false
			}
			return true
		})
	}
	return fd
}
