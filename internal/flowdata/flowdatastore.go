package flowdata

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
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

func (fds *FlowDataStore) GetFlowSums(filter FilterAttributes) []*FlowSum {
	fs := []*FlowSum{}
	useFilter := (filter != (FilterAttributes{}))
	if !useFilter {
		err := fds.db.AllByIndex("Key", &fs)
		if err != nil && !errors.Is(err, storm.ErrNotFound) {
			logrus.WithError(err).Panic("error getting all flow sums")
		}
	} else {
		matchers := []q.Matcher{}
		if filter.Action != "" {
			matchers = append(matchers, q.Eq("Action", filter.Action))
		}
		if filter.Port > 0 {
			matchers = append(matchers, q.Eq("DestPort", filter.Port))
		}
		if filter.Namespace != "" {
			qor := q.Or(
				q.Re("SourceNamespace", filter.Namespace),
				q.Re("DestNamespace", filter.Namespace),
			)
			matchers = append(matchers, qor)
		}
		if filter.Name != "" {
			qor := q.Or(
				q.Re("SourceName", filter.Name),
				q.Re("DestName", filter.Name),
			)
			matchers = append(matchers, qor)
		}
		query := fds.db.Select(matchers...).OrderBy("Key")
		err := query.Find(&fs)
		if err != nil && !errors.Is(err, storm.ErrNotFound) {
			logrus.WithError(err).Panic("error getting all flow sums")
		}
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
	if filter.Action != "" {
		fd = util.FilterSlice(fd, func(f *FlowData) bool {
			return f.Action == filter.Action
		})
	}
	return fd
}
