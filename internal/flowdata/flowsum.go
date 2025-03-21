package flowdata

import (
	"errors"
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

// flowdata.FlowItem interface
func (fs *FlowSum) GetID() int {
	return fs.ID
}

func toSum(fd *FlowData, fs *FlowSum) *FlowSum {
	if fs == nil {
		fs = &FlowSum{}
		fs.Key = fd.SumKey()
		fs.StartTime = fd.StartTime
		fs.EndTime = fd.EndTime
	} else {
		fs.StartTime = util.MinTime(fs.StartTime, fd.StartTime)
		fs.EndTime = util.MaxTime(fs.EndTime, fd.EndTime)
	}
	fs.Action = fd.Action
	fs.SourceName = fd.SourceName
	fs.SourceNamespace = fd.SourceNamespace
	fs.SourceLabels = fd.SourceLabels
	fs.DestName = fd.DestName
	fs.DestNamespace = fd.DestNamespace
	fs.DestLabels = fd.DestLabels
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
