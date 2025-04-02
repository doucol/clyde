package flowcache

import (
	"context"
	"fmt"
	"time"

	"github.com/doucol/clyde/internal/cache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
)

type FlowCache struct {
	fds          *flowdata.FlowDataStore
	flowSumCache *cache.Cache[string, []*flowdata.FlowSum]
	flowCache    *cache.Cache[string, []*flowdata.FlowData]
}

const (
	flowSumCacheName = "flowSums"
	flowDataBySumID  = "flowsBySumID"
)

func NewFlowCache(ctx context.Context, fds *flowdata.FlowDataStore) *FlowCache {
	fc := &FlowCache{
		fds:          fds,
		flowSumCache: cache.New[string, []*flowdata.FlowSum](),
		flowCache:    cache.New[string, []*flowdata.FlowData](),
	}
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
	if flowSums, ok := fc.flowSumCache.Get(flowSumCacheName); ok {
		return flowSums
	}
	return []*flowdata.FlowSum{}
}

func (fc *FlowCache) GetFlowsBySumID(sumID int) []*flowdata.FlowData {
	key := fmt.Sprintf("%s-%d", flowDataBySumID, sumID)
	if flows, ok := fc.flowCache.Get(key); ok || len(flows) > 0 {
		if !ok {
			go fc.cacheFlowsBySumID(key, sumID)
		}
		return flows
	}
	return fc.cacheFlowsBySumID(key, sumID)
}

func (fc *FlowCache) refreshCache() {
	fc.cacheFlowSums()
}

func (fc *FlowCache) cacheFlowsBySumID(key string, sumID int) []*flowdata.FlowData {
	flows := fc.fds.GetFlowsBySumID(sumID, global.GetFilter())
	fc.flowCache.SetTTL(key, flows, 5*time.Second)
	return flows
}

func (fc *FlowCache) cacheFlowSums() []*flowdata.FlowSum {
	flowSums := fc.fds.GetFlowSums(global.GetFilter())
	fc.flowSumCache.Set(flowSumCacheName, flowSums)
	return flowSums
}
