package global

import (
	"sync"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/rivo/tview"
)

type GlobalState struct {
	Filter flowdata.FilterAttributes
	app    *tview.Application
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
