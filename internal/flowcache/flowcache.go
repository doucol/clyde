package flowcache

import (
	"context"
	"time"

	"github.com/doucol/clyde/internal/cache"
	"github.com/doucol/clyde/internal/flowdata"
)

type flowCacheEntry []flowdata.FlowItem

type FlowCache struct {
	fds *flowdata.FlowDataStore
	c   *cache.Cache[string, flowCacheEntry]
}

const flowSumCacheName = "flowSums"

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

func (fc *FlowCache) refreshCache() {
	fc.cacheFlowSums()
}

func (fc *FlowCache) cacheFlowSums() flowCacheEntry {
	afs := fc.fds.GetAllFlowSums()
	flowSums := make(flowCacheEntry, len(afs))
	for i, fs := range afs {
		flowSums[i] = fs
	}
	fc.c.Set(flowSumCacheName, flowSums)
	return flowSums
}
