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

func ChanWaitTimeout[T any](cWait chan T, seconds time.Duration, cSignal ...chan T) error {
	select {
	case <-cWait:
		ChanClose(cSignal...)
		return nil
	case <-time.After(time.Second * seconds):
		return errors.New("timeout waiting for channel signal")
	}
}
