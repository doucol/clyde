package stopper

import (
	"fmt"
	"sync"
)

type Stopper interface {
	Stop()
}

type Closer interface {
	Close()
}

var (
	mu       = sync.Mutex{}
	closers  = []Closer{}
	stoppers = []Stopper{}
)

func AddClosers(c ...Closer) {
	mu.Lock()
	defer mu.Unlock()
	closers = append(closers, c...)
}

func AddStoppers(s ...Stopper) {
	mu.Lock()
	defer mu.Unlock()
	stoppers = append(stoppers, s...)
}

func Panic(err error) {
	PanicMsg("panic error occurred", err)
}

func PanicMsg(msg string, err error) {
	mu.Lock()
	defer mu.Unlock()
	defer panic(fmt.Errorf("%s: %w", msg, err))
	for _, s := range stoppers {
		s.Stop()
	}
	for _, c := range closers {
		c.Close()
	}
}
