package flowdata

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/doucol/clyde/internal/util"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
)

type FlowDataStore struct {
	db     *storm.DB
	inFlow chan Flower
	wg     sync.WaitGroup
	stop   chan struct{}
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
	ic := make(chan Flower, 1000)
	return &FlowDataStore{db: db, inFlow: ic, wg: sync.WaitGroup{}, stop: make(chan struct{})}, nil
}

func Clear() error {
	dbPath := dbPath()
	if util.FileExists(dbPath) {
		return os.Remove(dbPath)
	}
	return nil
}

func (fds *FlowDataStore) Run(recoverFunc func()) {
	fds.wg.Add(1)
	go func() {
		defer fds.wg.Done()
		if recoverFunc != nil {
			defer recoverFunc()
		}
		for f := range fds.inFlow {
			if f == nil {
				return
			}
			switch fl := f.(type) {
			case *FlowData:
				fs, newSum, err := fds.addFlow(fl)
				if err != nil {
					panic(err)
				}
				if newSum {
					logrus.Tracef("added flow data: new flow sum: %s", fs.Key)
				} else {
					logrus.Tracef("added flow data: existing flow sum: %s", fs.Key)
				}
			case *FlowSum:
				if err := fds.db.Save(fl); err != nil {
					logrus.WithError(err).Panic("error saving flow sum")
				} else {
					logrus.Tracef("updated flow sum: %s", fl.Key)
				}
			default:
				panic("unknown type in inFlow channel")
			}
		}
	}()

	fds.wg.Add(1)
	go func() {
		defer fds.wg.Done()
		if recoverFunc != nil {
			defer recoverFunc()
		}
		// TODO: add ability to configure the window
		epoch := time.Now().UTC()
		window := time.Minute * 5
		tock := time.Tick(5 * time.Second)
		for {
			select {
			case <-tock:
				fds.calcRates(epoch, window)
			case <-fds.stop:
				return
			}
		}
	}()
	return
}

func (fds *FlowDataStore) Close() {
	// send stop signals
	fds.stop <- struct{}{}
	fds.inFlow <- nil
	// wait for goroutines to finish
	fds.wg.Wait()
	// and close everything
	close(fds.stop)
	close(fds.inFlow)
	runtime.HandleError(fds.db.Close())
}

func (fds *FlowDataStore) addFlow(fd *FlowData) (*FlowSum, bool, error) {
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

func (fds *FlowDataStore) AddFlow(fd *FlowData) {
	fds.inFlow <- fd
}

func (fds *FlowDataStore) calcRates(epoch time.Time, window time.Duration) {
	secondsSinceEpoch := time.Now().UTC().Sub(epoch).Seconds()
	windowSeconds := window.Seconds()
	rateSeconds := min(secondsSinceEpoch, windowSeconds)
	durationToSubtract := time.Duration(rateSeconds * float64(time.Second) * -1)

	fss := fds.GetFlowSums(FilterAttributes{})
	for _, fs := range fss {
		var srcPacketsInSum, srcPacketsOutSum, srcBytesInSum, srcBytesOutSum uint64
		var dstPacketsInSum, dstPacketsOutSum, dstBytesInSum, dstBytesOutSum uint64
		fa := FilterAttributes{DateFrom: time.Now().UTC().Add(durationToSubtract)}
		for _, fd := range fds.GetFlowsBySumID(fs.ID, fa) {
			reporter := strings.ToLower(fd.Reporter)
			switch reporter {
			case "src":
				srcPacketsInSum += uint64(fd.PacketsIn)
				srcPacketsOutSum += uint64(fd.PacketsOut)
				srcBytesInSum += uint64(fd.BytesIn)
				srcBytesOutSum += uint64(fd.BytesOut)
			case "dst":
				dstPacketsInSum += uint64(fd.PacketsIn)
				dstPacketsOutSum += uint64(fd.PacketsOut)
				dstBytesInSum += uint64(fd.BytesIn)
				dstBytesOutSum += uint64(fd.BytesOut)
			}
		}

		fs.SourcePacketsInRate = float64(srcPacketsInSum) / rateSeconds
		fs.SourcePacketsOutRate = float64(srcPacketsOutSum) / rateSeconds
		fs.SourceBytesInRate = float64(srcBytesInSum) / rateSeconds
		fs.SourceBytesOutRate = float64(srcBytesOutSum) / rateSeconds

		fs.DestPacketsInRate = float64(dstPacketsInSum) / rateSeconds
		fs.DestPacketsOutRate = float64(dstPacketsOutSum) / rateSeconds
		fs.DestBytesInRate = float64(dstBytesInSum) / rateSeconds
		fs.DestBytesOutRate = float64(dstBytesOutSum) / rateSeconds

		fs.SourceTotalPacketRate = float64(srcPacketsInSum+srcPacketsOutSum) / rateSeconds
		fs.SourceTotalByteRate = float64(srcBytesInSum+srcBytesOutSum) / rateSeconds
		fs.DestTotalPacketRate = float64(dstPacketsInSum+dstPacketsOutSum) / rateSeconds
		fs.DestTotalByteRate = float64(dstBytesInSum+dstBytesOutSum) / rateSeconds
		fds.inFlow <- fs
	}
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
