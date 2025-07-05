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
	wg     *sync.WaitGroup
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
	return &FlowDataStore{
		db:     db,
		stop:   make(chan struct{}),
		inFlow: make(chan Flower, 1000),
	}, nil
}

func Clear() error {
	dbPath := dbPath()
	if util.FileExists(dbPath) {
		return os.Remove(dbPath)
	}
	return nil
}

func (fds *FlowDataStore) Run(recoverFunc func()) {
	fds.wg = &sync.WaitGroup{}
	fds.wg.Add(1)
	go func() {
		defer fds.wg.Done()
		if recoverFunc != nil {
			defer recoverFunc()
		}
		for f := range fds.inFlow {
			if f == nil {
				logrus.Debug("stop signal received, exiting inFlow goroutine")
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
				logrus.Debug("stop signal received, exiting rate calculation")
				return
			}
		}
	}()
}

func (fds *FlowDataStore) Close() {
	logrus.Debug("closing flow data store")
	if fds.wg != nil {
		// send stop signals
		fds.stop <- struct{}{}
		fds.inFlow <- nil
		// wait for goroutines to finish
		fds.wg.Wait()
	}
	// and close everything
	close(fds.stop)
	close(fds.inFlow)
	if err := fds.db.Close(); err != nil {
		logrus.WithError(err).Error("error closing flow data store")
	}
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
	const year = time.Hour * 24 * 365
	now := time.Now().UTC()
	secondsSinceEpoch := now.Sub(epoch).Seconds()
	windowSeconds := window.Seconds()
	minSeconds := min(secondsSinceEpoch, windowSeconds)
	durationToSubtract := time.Duration(minSeconds * float64(time.Second) * -1)
	filter := FilterAttributes{DateFrom: now.Add(durationToSubtract)}
	startTime, endTime := now.Add(year), now.Add(-year)

	fss := fds.GetFlowSums(FilterAttributes{})

	for _, fs := range fss {
		var srcPacketsInSum, srcPacketsOutSum, srcBytesInSum, srcBytesOutSum uint64
		var dstPacketsInSum, dstPacketsOutSum, dstBytesInSum, dstBytesOutSum uint64
		srcStartTime, dstStartTime := startTime, startTime
		srcEndTime, dstEndTime := endTime, endTime

		for _, fd := range fds.GetFlowsBySumID(fs.ID, filter) {
			switch strings.ToLower(fd.Reporter) {
			case "src":
				srcStartTime = util.MinTime(fd.StartTime, srcStartTime)
				srcEndTime = util.MaxTime(fd.EndTime, srcEndTime)
				srcPacketsInSum += uint64(fd.PacketsIn)
				srcPacketsOutSum += uint64(fd.PacketsOut)
				srcBytesInSum += uint64(fd.BytesIn)
				srcBytesOutSum += uint64(fd.BytesOut)
			case "dst":
				dstStartTime = util.MinTime(fd.StartTime, dstStartTime)
				dstEndTime = util.MaxTime(fd.EndTime, dstEndTime)
				dstPacketsInSum += uint64(fd.PacketsIn)
				dstPacketsOutSum += uint64(fd.PacketsOut)
				dstBytesInSum += uint64(fd.BytesIn)
				dstBytesOutSum += uint64(fd.BytesOut)
			}
		}

		srcRateSeconds := max(srcEndTime.Sub(srcStartTime).Seconds(), 1)
		dstRateSeconds := max(dstEndTime.Sub(dstStartTime).Seconds(), 1)

		fs.SourcePacketsInRate = float64(srcPacketsInSum) / srcRateSeconds
		fs.SourcePacketsOutRate = float64(srcPacketsOutSum) / srcRateSeconds
		fs.SourceBytesInRate = float64(srcBytesInSum) / srcRateSeconds
		fs.SourceBytesOutRate = float64(srcBytesOutSum) / srcRateSeconds

		fs.DestPacketsInRate = float64(dstPacketsInSum) / dstRateSeconds
		fs.DestPacketsOutRate = float64(dstPacketsOutSum) / dstRateSeconds
		fs.DestBytesInRate = float64(dstBytesInSum) / dstRateSeconds
		fs.DestBytesOutRate = float64(dstBytesOutSum) / dstRateSeconds

		fs.SourceTotalPacketRate = float64(srcPacketsInSum+srcPacketsOutSum) / srcRateSeconds
		fs.SourceTotalByteRate = float64(srcBytesInSum+srcBytesOutSum) / srcRateSeconds
		fs.DestTotalPacketRate = float64(dstPacketsInSum+dstPacketsOutSum) / dstRateSeconds
		fs.DestTotalByteRate = float64(dstBytesInSum+dstBytesOutSum) / dstRateSeconds
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
