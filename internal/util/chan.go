package util

import (
	"errors"
	"time"
)

func ChanSendEmpty[T any](ch chan T, count int) {
	for range count {
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

func ChanClose[T any](ch ...chan T) {
	for _, c := range ch {
		if c == nil {
			continue
		}
		select {
		case _, ok := <-c:
			if !ok {
				// channel is already closed
				continue
			}
			close(c)
		default:
			close(c)
		}
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
