package flowcache

import (
	"context"
	"fmt"
	"time"

	"github.com/doucol/clyde/internal/cache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
)

type flowCacheEntry []flowdata.FlowItem

type FlowCache struct {
	fds *flowdata.FlowDataStore
	c   *cache.Cache[string, flowCacheEntry]
}

const (
	flowSumCacheName = "flowSums"
	flowDataBySumID  = "flowsBySumID"
)

func NewFlowCache(ctx context.Context, fds *flowdata.FlowDataStore) *FlowCache {
	fc := &FlowCache{fds: fds, c: cache.New[string, flowCacheEntry]()}
	// Go refresh the cache every 2 seconds
	go func() {
		ticker := time.Tick(2 * time.Second)
		for {
			fc.refreshCache()
			select {
			case <-ctx.Done():
				return
			case <-ticker:
				continue
			}
		}
	}()
	return fc
}

func (fc *FlowCache) GetFlowSums() []*flowdata.FlowSum {
	if flowSums, ok := fc.c.Get(flowSumCacheName); ok {
		fss := make([]*flowdata.FlowSum, len(flowSums))
		for i, fsi := range flowSums {
			fss[i] = fsi.(*flowdata.FlowSum)
		}
		return fss
	}
	return []*flowdata.FlowSum{}
}

func (fc *FlowCache) GetFlowsBySumID(sumID int) []*flowdata.FlowData {
	key := fmt.Sprintf("%s-%d", flowDataBySumID, sumID)
	if flows, ok := fc.c.Get(key); ok || len(flows) > 0 {
		fd := make([]*flowdata.FlowData, len(flows))
		for i, f := range flows {
			fd[i] = f.(*flowdata.FlowData)
		}
		if !ok {
			go fc.cacheFlowsBySumID(key, sumID)
		}
		return fd
	}
	fce := fc.cacheFlowsBySumID(key, sumID)
	flows := make([]*flowdata.FlowData, len(fce))
	for i, f := range fce {
		flows[i] = f.(*flowdata.FlowData)
	}
	return flows
}

func (fc *FlowCache) refreshCache() {
	fc.cacheFlowSums()
}

func (fc *FlowCache) cacheFlowsBySumID(key string, sumID int) flowCacheEntry {
	gs := global.GetState()
	af := fc.fds.GetFlowsBySumID(sumID, &gs.Filter)
	flows := make(flowCacheEntry, len(af))
	for i, f := range af {
		flows[i] = f
	}
	fc.c.SetTTL(key, flows, 2*time.Second)
	return flows
}

func (fc *FlowCache) cacheFlowSums() flowCacheEntry {
	gs := global.GetState()
	afs := fc.fds.GetFlowSums(&gs.Filter)
	flowSums := make(flowCacheEntry, len(afs))
	for i, fs := range afs {
		flowSums[i] = fs
	}
	fc.c.Set(flowSumCacheName, flowSums)
	return flowSums
}
