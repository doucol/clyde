// Package flowcache provides a caching layer for flow data and flow sums.
package flowcache

import (
	"context"
	"fmt"
	"time"

	"github.com/doucol/clyde/internal/cache"
	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/global"
	"github.com/doucol/clyde/internal/util"
)

type FlowDataStore interface {
	GetFlowSums(filter flowdata.FilterAttributes) []*flowdata.FlowSum
	GetFlowsBySumID(sumID int, filter flowdata.FilterAttributes) []*flowdata.FlowData
}

type FlowCache struct {
	fds          FlowDataStore
	flowSumCache *cache.Cache[string, []*flowdata.FlowSum]
	flowCache    *cache.Cache[string, []*flowdata.FlowData]
}

const (
	flowSumCacheName = "flowSums"
	flowDataBySumID  = "flowsBySumID"
)

func NewFlowCache(ctx context.Context, fds FlowDataStore) *FlowCache {
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

// filterTag returns a stable cache-key suffix for a filter. The empty filter
// yields no suffix so the common (unfiltered) case keeps a stable base key.
// Including the filter in every cache key ensures a filter change produces a
// cache miss and fresh results, rather than serving stale entries until the
// next refresh/TTL.
func filterTag(f flowdata.FilterAttributes) string {
	if f == (flowdata.FilterAttributes{}) {
		return ""
	}
	return fmt.Sprintf("|f=%v", f)
}

func (fc *FlowCache) baseKey(filter flowdata.FilterAttributes) string {
	return flowSumCacheName + filterTag(filter)
}

func (fc *FlowCache) cacheSortedFlowSums(cacheKey, fieldName string, ascending bool, filter flowdata.FilterAttributes) []*flowdata.FlowSum {
	flowSums, _ := fc.flowSumCache.Get(fc.baseKey(filter))
	if flowSums == nil {
		flowSums = fc.cacheFlowSums()
	}
	if len(flowSums) > 0 {
		fsc := make([]*flowdata.FlowSum, len(flowSums))
		copy(fsc, flowSums)
		util.SortSlice(fsc, fieldName, ascending)
		fc.flowSumCache.SetTTL(cacheKey, fsc, time.Second*2)
		return fsc
	}
	return []*flowdata.FlowSum{}
}

func (fc *FlowCache) getFlowSums(sortBy string, asc bool) []*flowdata.FlowSum {
	filter := global.GetFilter()
	cacheKey := fc.baseKey(filter)
	if sortBy != "" {
		cacheKey = fmt.Sprintf("%s-%s-%t", cacheKey, sortBy, asc)
	}
	if flowSums, ok := fc.flowSumCache.Get(cacheKey); ok || len(flowSums) > 0 {
		if !ok && sortBy != "" {
			go fc.cacheSortedFlowSums(cacheKey, sortBy, asc, filter)
		}
		return flowSums
	} else if sortBy != "" {
		return fc.cacheSortedFlowSums(cacheKey, sortBy, asc, filter)
	}
	return fc.cacheFlowSums()
}

func (fc *FlowCache) GetFlowSumTotals() []*flowdata.FlowSum {
	sa := global.GetSort()
	return fc.getFlowSums(sa.SumTotalsFieldName, sa.SumTotalsAscending)
}

func (fc *FlowCache) GetFlowSumRates() []*flowdata.FlowSum {
	sa := global.GetSort()
	return fc.getFlowSums(sa.SumRatesFieldName, sa.SumRatesAscending)
}

func (fc *FlowCache) GetFlowsBySumID(sumID int) []*flowdata.FlowData {
	filter := global.GetFilter()
	key := fmt.Sprintf("%s-%d%s", flowDataBySumID, sumID, filterTag(filter))
	if flows, ok := fc.flowCache.Get(key); ok || len(flows) > 0 {
		if !ok {
			go fc.cacheFlowsBySumID(key, sumID, filter)
		}
		return flows
	}
	return fc.cacheFlowsBySumID(key, sumID, filter)
}

func (fc *FlowCache) refreshCache() {
	fc.cacheFlowSums()
}

func (fc *FlowCache) cacheFlowsBySumID(key string, sumID int, filter flowdata.FilterAttributes) []*flowdata.FlowData {
	flows := fc.fds.GetFlowsBySumID(sumID, filter)
	fc.flowCache.SetTTL(key, flows, 5*time.Second)
	return flows
}

func (fc *FlowCache) cacheFlowSums() []*flowdata.FlowSum {
	filter := global.GetFilter()
	flowSums := fc.fds.GetFlowSums(filter)
	// TTL (rather than forever) so cached sets for filters no longer in use
	// are eventually culled; the 2s refresh loop keeps the active set warm.
	fc.flowSumCache.SetTTL(fc.baseKey(filter), flowSums, 10*time.Second)
	return flowSums
}
