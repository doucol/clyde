package util

func ChanSendEmpty[T any](ch chan T, count int) {
    for i := 0; i < count; i++ {
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
