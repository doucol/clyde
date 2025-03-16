package util

func ChanSendZeroedVals[T any](ch chan T, count int) {
	for range count {
		var val T
		ch <- val
	}
}

func ChanClose[T any](ch ...chan T) {
	for _, c := range ch {
		select {
		case _, ok := <-c:
			if ok {
				close(c)
			}
		default:
			close(c)
		}
	}
}
