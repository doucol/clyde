package flowdata

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	"github.com/blugelabs/bluge"
	"github.com/doucol/clyde/internal/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type FlowDataStore struct {
	db *storm.DB
	// idx bleve.Index
	idx *bluge.Writer
}

type FlowItem interface {
	GetID() int
}

func dbPath() string {
	return filepath.Join(util.GetDataPath(), "flowdata.db")
}

func idxPath() string {
	return filepath.Join(util.GetDataPath(), "flowdata.idx")
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

	config := bluge.DefaultConfig(idxPath())
	idx, err := bluge.OpenWriter(config)
	if err != nil {
		return nil, err
	}

	// doc := bluge.NewDocument("example").
	// 	AddField(bluge.NewTextField("name", "bluge"))
	//
	// err = writer.Update(doc.ID(), doc)
	// if err != nil {
	// 	return nil, err
	// }
	// var idx bleve.Index
	// idxpath := idxPath()
	// if util.FileExists(idxpath) {
	// 	idx, err = bleve.Open(idxpath)
	// } else {
	// 	mapping := bleve.NewIndexMapping()
	// 	idx, err = bleve.NewUsing(idxpath, mapping, "scorch", "scorch", nil)
	// }
	// if err != nil {
	// 	return nil, err
	// }
	// return &FlowDataStore{db: db, idx: idx}, nil
	return &FlowDataStore{db: db, idx: idx}, nil
}

func Clear() error {
	dbPath := dbPath()
	if util.FileExists(dbPath) {
		return os.Remove(dbPath)
	}
	idxPath := idxPath()
	if util.FileExists(idxPath) {
		return os.RemoveAll(idxPath)
	}
	return nil
}

func (fds *FlowDataStore) Close() {
	runtime.HandleError(fds.db.Close())
	runtime.HandleError(fds.idx.Close())
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
	err = fds.Index(fd)
	if err != nil {
		return nil, false, err
	}
	committed = true
	return fs, newSum, nil
}

func (fds *FlowDataStore) Index(fd *FlowData) error {
	doc := bluge.NewDocument(strconv.Itoa(fd.GetID()))
	doc.AddField(bluge.NewKeywordField("sum_id", strconv.Itoa(fd.SumID)))
	doc.AddField(bluge.NewTextField("source_namespace", fd.SourceNamespace))
	doc.AddField(bluge.NewTextField("source_name", fd.SourceName))
	doc.AddField(bluge.NewTextField("source_labels", fd.SourceLabels))
	doc.AddField(bluge.NewTextField("dest_namespace", fd.DestNamespace))
	doc.AddField(bluge.NewTextField("dest_name", fd.DestName))
	doc.AddField(bluge.NewTextField("dest_labels", fd.DestLabels))
	doc.AddField(bluge.NewKeywordField("proto", fd.Protocol))
	doc.AddField(bluge.NewKeywordField("dest_port", strconv.FormatInt(fd.DestPort, 10)))
	doc.AddField(bluge.NewKeywordField("action", fd.Action))
	doc.AddField(bluge.NewKeywordField("reporter", fd.Reporter))
	doc.AddField(bluge.NewDateTimeField("start_time", fd.StartTime))
	doc.AddField(bluge.NewDateTimeField("end_time", fd.EndTime))
	err := fds.idx.Update(doc.ID(), doc)
	return err
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
