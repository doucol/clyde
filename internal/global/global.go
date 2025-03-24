package global

import (
	"sync"

	"github.com/doucol/clyde/internal/flowdata"
)

type GlobalState struct {
	Filter flowdata.FilterAttributes
}

var (
	_mu = &sync.RWMutex{}
	_gs = &GlobalState{}
)

func GetState() GlobalState {
	_mu.RLock()
	defer _mu.RUnlock()
	return *_gs
}

func SetState(gs GlobalState) {
	_mu.Lock()
	defer _mu.Unlock()
	_gs = &gs
}
