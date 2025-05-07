package global

import (
	"sync"

	"github.com/doucol/clyde/internal/flowdata"
)

type GlobalState struct {
	Filter flowdata.FilterAttributes
	Sort   flowdata.SortAttributes
}

var (
	mu = sync.RWMutex{}
	gs = GlobalState{}
)

func GetState() GlobalState {
	mu.RLock()
	defer mu.RUnlock()
	return gs
}

func SetState(g GlobalState) {
	mu.Lock()
	defer mu.Unlock()
	gs = g
}

func GetFilter() flowdata.FilterAttributes {
	return GetState().Filter
}

func SetFilter(fa flowdata.FilterAttributes) {
	mu.Lock()
	defer mu.Unlock()
	gs.Filter = fa
}

func GetSort() flowdata.SortAttributes {
	return GetState().Sort
}

func SetSort(sort flowdata.SortAttributes) {
	mu.Lock()
	defer mu.Unlock()
	gs.Sort = sort
}
