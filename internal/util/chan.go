package util

import (
	"errors"
	"time"
)

func ChanSendEmpty[T any](ch chan T, count int) {
    for i := 0; i < count; i++ {
        var val T
        ch <- val
    }
}

func ChanSendTimeout[T any](ch chan T, val T, milliseconds int) error {
	if cap(ch) > 0 {
		panic("channel must be unbuffered")
	}
	select {
	case ch <- val:
		return nil
	case <-time.After(time.Millisecond * time.Duration(milliseconds)):
		return errors.New("timeout sending to channel")
	}
}

// ChanClose closes each non-nil channel. It is safe to call on a channel that
// is already closed: the resulting "close of closed channel" panic is
// recovered, making the close idempotent. Callers must still ensure no
// goroutine is concurrently sending on or closing the same channel.
func ChanClose[T any](ch ...chan T) {
	for _, c := range ch {
		if c == nil {
			continue
		}
		func() {
			defer func() { _ = recover() }()
			close(c)
		}()
	}
}

func ChanWaitTimeout[T any](cWait chan T, seconds time.Duration, cSignal ...chan T) (T, error) {
	select {
	case v := <-cWait:
		ChanClose(cSignal...)
		return v, nil
	case <-time.After(time.Second * seconds):
		var empty T
		return empty, errors.New("timeout waiting for channel signal")
	}
}
