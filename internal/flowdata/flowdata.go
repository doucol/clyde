package flowdata

import (
	"fmt"
	"time"

	storm "github.com/asdine/storm/v3"
)

type FlowResponse struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Action          string    `json:"action"`
	SourceName      string    `json:"source_name"`
	SourceNamespace string    `json:"source_namespace"`
	SourceLabels    string    `json:"source_labels"`
	DestName        string    `json:"dest_name"`
	DestNamespace   string    `json:"dest_namespace"`
	DestLabels      string    `json:"dest_labels"`
	Protocol        string    `json:"protocol"`
	DestPort        int64     `json:"dest_port"`
	Reporter        string    `json:"reporter"`
	PacketsIn       int64     `json:"packets_in"`
	PacketsOut      int64     `json:"packets_out"`
	BytesIn         int64     `json:"bytes_in"`
	BytesOut        int64     `json:"bytes_out"`
}

type FlowData struct {
	ID           int `storm:"id,increment"`
	FlowResponse `storm:"inline"`
}

type FlowDataStore struct {
	db *storm.DB
}

func NewFlowDataStore() (*FlowDataStore, error) {
	db, err := storm.Open("/tmp/flowdata.db")
	if err != nil {
		return nil, err
	}
	err = db.Init(&FlowData{})
	if err != nil {
		return nil, err
	}
	return &FlowDataStore{db: db}, nil
}

func (fds *FlowDataStore) Close() {
	err := fds.db.Close()
	if err != nil {
		panic(err)
	}
}

func (fds *FlowDataStore) Add(fd *FlowData) error {
	return fds.db.Save(fd)
}

func (fds *FlowDataStore) Get(id int) (*FlowData, bool) {
	fd := &FlowData{}
	err := fds.db.One("ID", id, fd)
	if err != nil {
		return nil, false
	}
	return fd, true
}

func (fds *FlowDataStore) Count() int {
	cnt, err := fds.db.Count(&FlowData{})
	if err != nil {
		fmt.Println(fmt.Errorf("error counting flow data: %v", err))
		panic(err)
	}
	return cnt
}
