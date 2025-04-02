package flowdata

import (
	"errors"
	"strings"
	"time"

	"github.com/doucol/clyde/internal/util"
)

type FlowSum struct {
	ID               int       `json:"id" storm:"id,increment"`
	Key              string    `json:"key" storm:"unique"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	Action           string    `json:"action"`
	SourceName       string    `json:"source_name"`
	SourceNamespace  string    `json:"source_namespace"`
	SourceLabels     string    `json:"source_labels"`
	DestName         string    `json:"dest_name"`
	DestNamespace    string    `json:"dest_namespace"`
	DestLabels       string    `json:"dest_labels"`
	Protocol         string    `json:"protocol"`
	DestPort         int64     `json:"dest_port"`
	SourceReports    int64     `json:"source_reports"`
	DestReports      int64     `json:"dest_reports"`
	SourcePacketsIn  uint64    `json:"source_packets_in"`
	SourcePacketsOut uint64    `json:"source_packets_out"`
	SourceBytesIn    uint64    `json:"source_bytes_in"`
	SourceBytesOut   uint64    `json:"source_bytes_out"`
	DestPacketsIn    uint64    `json:"dest_packets_in"`
	DestPacketsOut   uint64    `json:"dest_packets_out"`
	DestBytesIn      uint64    `json:"dest_bytes_in"`
	DestBytesOut     uint64    `json:"dest_bytes_out"`
}

// [Flower] interface
func (fs *FlowSum) GetID() int {
	return fs.ID
}

func (fs *FlowSum) GetSumKey() string {
	return fs.Key
}

func (fs *FlowSum) GetSourceNamespace() string {
	return fs.SourceNamespace
}

func (fs *FlowSum) GetSourceName() string {
	return fs.SourceName
}

func (fs *FlowSum) GetSourceLabels() string {
	return fs.SourceLabels
}

func (fs *FlowSum) GetDestNamespace() string {
	return fs.DestNamespace
}

func (fs *FlowSum) GetDestName() string {
	return fs.DestName
}

func (fs *FlowSum) GetDestLabels() string {
	return fs.DestLabels
}

func (fs *FlowSum) GetAction() string {
	return fs.Action
}

func (fs *FlowSum) GetPort() int64 {
	return fs.DestPort
}

func (fs *FlowSum) GetStartTime() time.Time {
	return fs.StartTime
}

func (fs *FlowSum) GetEndTime() time.Time {
	return fs.EndTime
}

func flowToFlowSum(fd *FlowData, fs *FlowSum) *FlowSum {
	if fs == nil {
		fs = &FlowSum{}
		fs.Key = fd.GetSumKey()
		fs.StartTime = fd.StartTime
		fs.EndTime = fd.EndTime
	} else {
		fs.StartTime = util.MinTime(fs.StartTime, fd.StartTime)
		fs.EndTime = util.MaxTime(fs.EndTime, fd.EndTime)
	}
	fs.Action = fd.Action
	fs.SourceName = fd.SourceName
	fs.SourceNamespace = fd.SourceNamespace
	fs.SourceLabels = aggregateLabels(fd.SourceLabels, fs.SourceLabels)
	fs.DestName = fd.DestName
	fs.DestNamespace = fd.DestNamespace
	fs.DestLabels = aggregateLabels(fd.DestLabels, fs.DestLabels)
	fs.Protocol = fd.Protocol
	fs.DestPort = fd.DestPort
	switch fd.Reporter {
	case Reporter_name[int32(Reporter_Src)]:
		fs.SourceReports += 1
		fs.SourcePacketsIn += uint64(fd.PacketsIn)
		fs.SourcePacketsOut += uint64(fd.PacketsOut)
		fs.SourceBytesIn += uint64(fd.BytesIn)
		fs.SourceBytesOut += uint64(fd.BytesOut)
	case Reporter_name[int32(Reporter_Dst)]:
		fs.DestReports += 1
		fs.DestPacketsIn += uint64(fd.PacketsIn)
		fs.DestPacketsOut += uint64(fd.PacketsOut)
		fs.DestBytesIn += uint64(fd.BytesIn)
		fs.DestBytesOut += uint64(fd.BytesOut)
	default:
		panic(errors.New("unknown reporter in flow data: " + Reporter_name[int32(Reporter_value[fd.Reporter])]))
	}
	return fs
}

func aggregateLabels(flowLabels string, sumLabels string) string {
	sumSlice := strings.Split(sumLabels, "|")
	set := make(map[string]any, len(sumSlice))
	for _, item := range sumSlice {
		set[strings.TrimSpace(item)] = struct{}{}
	}
	flowSlice := strings.Split(flowLabels, "|")
	for _, item := range flowSlice {
		set[strings.TrimSpace(item)] = struct{}{}
	}
	return util.JoinMapKeys(set, " | ")
}
