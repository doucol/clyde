package flowdata

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/asdine/storm/v3"
	"github.com/doucol/clyde/internal/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
)

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
	fs = toSum(fd, fs)
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

func (fds *FlowDataStore) GetAllFlowSums() []*FlowSum {
	fs := []*FlowSum{}
	err := fds.db.All(&fs)
	if err != nil {
		logrus.WithError(err).Panic("error getting all flow sums")
	}
	return fs
}

func (fds *FlowDataStore) GetFlowSumCount() int {
	cnt, err := fds.db.Count(&FlowSum{})
	if err != nil {
		logrus.WithError(err).Panic("error getting flow sum count")
	}
	return cnt
}

func (fds *FlowDataStore) GetFlowDetail(key string, idx int) (*FlowData, bool) {
	fs := &FlowSum{}
	err := fds.db.One("Key", key, fs)
	if err != nil {
		panic(fmt.Errorf("error getting flow aggregate: %v", err))
	}

	fd := []FlowData{}
	err = fds.db.Find("SumID", fs.ID, &fd, storm.Skip(idx-1), storm.Limit(1))
	if err != nil {
		panic(fmt.Errorf("error getting flow detail: %s, %d, %v", key, idx, fs))
	}
	return &fd[0], true
}

func (fds *FlowDataStore) GetAllFlowsByKey(key string) []*FlowData {
	fs := &FlowSum{}
	err := fds.db.One("Key", key, fs)
	if err != nil {
		panic(fmt.Errorf("error getting flow aggregate: %v", err))
	}

	fd := []*FlowData{}
	err = fds.db.Find("SumID", fs.ID, &fd)
	if err != nil {
		panic(fmt.Errorf("error getting flow detail: %s, %v", key, fs))
	}
	return fd
}

func (fds *FlowDataStore) GetFlowDetailCount(key string) int {
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
