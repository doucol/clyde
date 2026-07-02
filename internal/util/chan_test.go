package util

import (
	"testing"
	"time"
)

func TestChanSendEmpty(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected int
	}{
		{
			name:     "send zero values",
			count:    0,
			expected: 0,
		},
		{
			name:     "send one value",
			count:    1,
			expected: 1,
		},
		{
			name:     "send multiple values",
			count:    5,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan int, tt.count)
			ChanSendEmpty(ch, tt.count)

			// Check if the correct number of values were sent
			received := 0
			for range tt.count {
				select {
				case <-ch:
					received++
				case <-time.After(100 * time.Millisecond):
					t.Errorf("Timeout waiting for value")
					return
				}
			}

			if received != tt.expected {
				t.Errorf("ChanSendEmpty() sent %v values; want %v", received, tt.expected)
			}
		})
	}
}

func TestChanClose(t *testing.T) {
	tests := []struct {
		name     string
		channels int
	}{
		{
			name:     "close single channel",
			channels: 1,
		},
		{
			name:     "close multiple channels",
			channels: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test channels
			channels := make([]chan int, tt.channels)
			for i := range channels {
				channels[i] = make(chan int)
			}

			// Close the channels
			ChanClose(channels...)

			// Verify all channels are closed
			for i, ch := range channels {
				select {
				case _, ok := <-ch:
					if ok {
						t.Errorf("Channel %d is not closed", i)
					}
				default:
					t.Errorf("Channel %d is not closed", i)
				}
			}

			// Closing again must be a safe no-op (idempotent), not a panic.
			ChanClose(channels...)
		})
	}
}

func TestChanCloseIdempotentAndNilSafe(t *testing.T) {
	// A nil channel in the list is skipped.
	var nilCh chan int
	ch := make(chan int)
	ChanClose(nilCh, ch)

	// Repeated close of the same channel does not panic.
	ChanClose(ch)
	ChanClose(ch, ch)

	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed")
		}
	default:
		t.Error("expected channel to be closed")
	}
}
