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

	// The event channels below are lazily created by their accessors and are
	// read by the consumer goroutine, so access is guarded by chanMu. They are
	// used purely as observation hooks (currently only by tests); in normal
	// operation no accessor is called, the fields stay nil, and chanSignal is a
	// cheap no-op.
	chanMu           sync.RWMutex
	flowAdded        chan Flower
	flowSumAdded     chan Flower
	flowSumsUpdated  chan Flower
	flowRatesUpdated chan Flower

	RateCalcWindow   int
	RateCalcInterval int
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
		db:               db,
		stop:             make(chan struct{}, 1),
		inFlow:           make(chan Flower, 1000),
		RateCalcWindow:   60, // Default to 60 seconds
		RateCalcInterval: 5,  // Default to 5 seconds
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
		for {
			select {
			case <-fds.stop:
				return
			case f := <-fds.inFlow:
				switch fl := f.(type) {
				case *FlowData:
					fs, newSum, err := fds.addFlow(fl)
					if err != nil {
						logrus.WithError(err).Error("error adding flow; skipping")
						continue
					}
					fds.chanMu.RLock()
					cAdded, cSumAdded, cSumsUpdated := fds.flowAdded, fds.flowSumAdded, fds.flowSumsUpdated
					fds.chanMu.RUnlock()
					chanSignal(cAdded, f)
					if newSum {
						chanSignal(cSumAdded, f)
						logrus.Tracef("added flow data: new flow sum: %s", fs.Key)
					} else {
						chanSignal(cSumsUpdated, f)
						logrus.Tracef("added flow data: existing flow sum: %s", fs.Key)
					}
				case *FlowSum:
					fds.chanMu.RLock()
					cRates := fds.flowRatesUpdated
					fds.chanMu.RUnlock()
					if err := fds.updateRates(fl); err != nil {
						logrus.WithError(err).Error("error updating flow sum rates")
					} else {
						chanSignal(cRates, f)
						logrus.Tracef("updated flow sum: %s", fl.Key)
					}
				default:
					panic("unknown type in inFlow channel")
				}
			}
		}
	}()

	fds.wg.Add(1)
	go func() {
		defer fds.wg.Done()
		if recoverFunc != nil {
			defer recoverFunc()
		}
		tock := time.Tick(time.Duration(fds.RateCalcInterval) * time.Second)
		for {
			select {
			case <-fds.stop:
				logrus.Debug("stop signal received, exiting rate calculation")
				return
			case <-tock:
				fds.calcRates()
			}
		}
	}()
}

func (fds *FlowDataStore) FlowAdded() chan Flower {
	fds.chanMu.Lock()
	defer fds.chanMu.Unlock()
	if fds.flowAdded == nil {
		fds.flowAdded = make(chan Flower)
	}
	return fds.flowAdded
}

func (fds *FlowDataStore) FlowSumAdded() chan Flower {
	fds.chanMu.Lock()
	defer fds.chanMu.Unlock()
	if fds.flowSumAdded == nil {
		fds.flowSumAdded = make(chan Flower)
	}
	return fds.flowSumAdded
}

func (fds *FlowDataStore) FlowSumsUpdated() chan Flower {
	fds.chanMu.Lock()
	defer fds.chanMu.Unlock()
	if fds.flowSumsUpdated == nil {
		fds.flowSumsUpdated = make(chan Flower)
	}
	return fds.flowSumsUpdated
}

func (fds *FlowDataStore) FlowRatesUpdated() chan Flower {
	fds.chanMu.Lock()
	defer fds.chanMu.Unlock()
	if fds.flowRatesUpdated == nil {
		fds.flowRatesUpdated = make(chan Flower)
	}
	return fds.flowRatesUpdated
}

func chanSignal[T any](ch chan T, val T) {
	if ch == nil {
		return
	}
	if err := util.ChanSendTimeout(ch, val, 10); err != nil {
		logrus.WithError(err).Error("error sending to channel")
	}
}

func (fds *FlowDataStore) Close() {
	logrus.Debug("closing flow data store")
	// Signal shutdown first so the producer (AddFlow) and rate goroutine stop
	// feeding inFlow, then wait for the store goroutines to drain before
	// closing the channels they read from.
	util.ChanClose(fds.stop)
	if fds.wg != nil {
		fds.wg.Wait()
	}
	util.ChanClose(fds.inFlow)
	fds.chanMu.RLock()
	added, sumAdded, sumsUpdated, ratesUpdated := fds.flowAdded, fds.flowSumAdded, fds.flowSumsUpdated, fds.flowRatesUpdated
	fds.chanMu.RUnlock()
	util.ChanClose(added, sumAdded, sumsUpdated, ratesUpdated)
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
	fs, err = flowToFlowSum(fd, fs)
	if err != nil {
		return nil, false, err
	}
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
	// Select on stop for both cases so a shutdown unblocks a send that would
	// otherwise wait on a full buffer, and so we never send on a closed inFlow.
	select {
	case <-fds.stop:
		return
	case fds.inFlow <- fd:
	}
}

func (fds *FlowDataStore) calcRates() {
	logrus.Debugf("calculating flow rates for window: %d", fds.RateCalcWindow)
	now := time.Now().UTC()
	const year = time.Hour * 24 * 365
	durationToSubtract := time.Duration(time.Second * time.Duration(-fds.RateCalcWindow))
	filter := FilterAttributes{DateFrom: now.Add(durationToSubtract)}
	startTime, endTime := now.Add(year), now.Add(-year)

	fss := fds.GetFlowSums(FilterAttributes{})

	logrus.Debugf("found %d flow sums to calculate rates for", len(fss))

	for _, fs := range fss {
		var srcPacketsInSum, srcPacketsOutSum, srcBytesInSum, srcBytesOutSum uint64
		var dstPacketsInSum, dstPacketsOutSum, dstBytesInSum, dstBytesOutSum uint64
		srcStartTime, dstStartTime := startTime, startTime
		srcEndTime, dstEndTime := endTime, endTime

		flowDataSet := fds.GetFlowsBySumID(fs.ID, filter)
		logrus.Tracef("processing %d flow data entries for filter %+v", len(flowDataSet), filter)

		for _, fd := range flowDataSet {
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
		logrus.Tracef("Source rates: PacketsInRate: %f, PacketsOutRate: %f, BytesInRate: %f, BytesOutRate: %f, sec: %f", fs.SourcePacketsInRate, fs.SourcePacketsOutRate, fs.SourceBytesInRate, fs.SourceBytesOutRate, srcRateSeconds)

		fs.DestPacketsInRate = float64(dstPacketsInSum) / dstRateSeconds
		fs.DestPacketsOutRate = float64(dstPacketsOutSum) / dstRateSeconds
		fs.DestBytesInRate = float64(dstBytesInSum) / dstRateSeconds
		fs.DestBytesOutRate = float64(dstBytesOutSum) / dstRateSeconds
		logrus.Tracef("Dest rates: PacketsInRate: %f, PacketsOutRate: %f, BytesInRate: %f, BytesOutRate: %f, sec: %f", fs.DestPacketsInRate, fs.DestPacketsOutRate, fs.DestBytesInRate, fs.DestBytesOutRate, dstRateSeconds)

		fs.SourceTotalPacketRate = float64(srcPacketsInSum+srcPacketsOutSum) / srcRateSeconds
		fs.SourceTotalByteRate = float64(srcBytesInSum+srcBytesOutSum) / srcRateSeconds
		logrus.Tracef("Total rates: SourceTotalPacketRate: %f, SourceTotalByteRate: %f, sec: %f", fs.SourceTotalPacketRate, fs.SourceTotalByteRate, srcRateSeconds)

		fs.DestTotalPacketRate = float64(dstPacketsInSum+dstPacketsOutSum) / dstRateSeconds
		fs.DestTotalByteRate = float64(dstBytesInSum+dstBytesOutSum) / dstRateSeconds
		logrus.Tracef("Total rates: DestTotalPacketRate: %f, DestTotalByteRate: %f, sec: %f", fs.DestTotalPacketRate, fs.DestTotalByteRate, dstRateSeconds)

		select {
		case <-fds.stop:
			return
		case fds.inFlow <- fs:
		}
	}
}

// updateRates merges the computed rate fields from a freshly calculated
// FlowSum into the current persisted record. Only the rate fields are written;
// the running totals and counters are owned by addFlow. Merging (rather than
// blindly saving the calculated copy) prevents a lost update when a flow for
// the same key is ingested between the rate calculation's read and this save.
func (fds *FlowDataStore) updateRates(rates *FlowSum) error {
	tx, err := fds.db.Begin(true)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			runtime.HandleError(tx.Rollback())
		}
	}()

	fs := &FlowSum{}
	if err := tx.One("ID", rates.ID, fs); err != nil {
		if errors.Is(err, storm.ErrNotFound) {
			// The sum was removed between rate calculation and this save.
			return nil
		}
		return err
	}

	fs.SourcePacketsInRate = rates.SourcePacketsInRate
	fs.SourcePacketsOutRate = rates.SourcePacketsOutRate
	fs.SourceBytesInRate = rates.SourceBytesInRate
	fs.SourceBytesOutRate = rates.SourceBytesOutRate
	fs.DestPacketsInRate = rates.DestPacketsInRate
	fs.DestPacketsOutRate = rates.DestPacketsOutRate
	fs.DestBytesInRate = rates.DestBytesInRate
	fs.DestBytesOutRate = rates.DestBytesOutRate
	fs.SourceTotalPacketRate = rates.SourceTotalPacketRate
	fs.SourceTotalByteRate = rates.SourceTotalByteRate
	fs.DestTotalPacketRate = rates.DestTotalPacketRate
	fs.DestTotalByteRate = rates.DestTotalByteRate

	if err := tx.Save(fs); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (fds *FlowDataStore) GetFlowSum(id int) *FlowSum {
	fs := &FlowSum{}
	err := fds.db.One("ID", id, fs)
	if err != nil {
		if !errors.Is(err, storm.ErrNotFound) {
			logrus.WithError(err).Error("error getting flow sum")
		}
		return nil
	}
	return fs
}

func (fds *FlowDataStore) GetFlowSums(filter FilterAttributes) []*FlowSum {
	fs := []*FlowSum{}
	err := fds.db.AllByIndex("Key", &fs)
	if err != nil && !errors.Is(err, storm.ErrNotFound) {
		logrus.WithError(err).Error("error getting all flow sums")
		return []*FlowSum{}
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
		if !errors.Is(err, storm.ErrNotFound) {
			logrus.WithError(err).Error("error getting flow data")
		}
		return nil
	}
	return fd
}

func (fds *FlowDataStore) GetFlowsBySumID(sumID int, filter FilterAttributes) []*FlowData {
	fd := []*FlowData{}
	err := fds.db.Find("SumID", sumID, &fd)
	if err != nil && !errors.Is(err, storm.ErrNotFound) {
		logrus.WithError(err).Error("error getting flows by sum id")
		return []*FlowData{}
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
