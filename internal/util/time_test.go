package util

import (
	"testing"
	"time"
)

func TestMinTime(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	oneHourLater := now.Add(1 * time.Hour)

	tests := []struct {
		name     string
		times    []time.Time
		expected time.Time
	}{
		{
			name:     "empty slice",
			times:    []time.Time{},
			expected: time.Time{},
		},
		{
			name:     "single time",
			times:    []time.Time{now},
			expected: now,
		},
		{
			name:     "multiple times",
			times:    []time.Time{now, oneHourAgo, twoHoursAgo, oneHourLater},
			expected: twoHoursAgo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MinTime(tt.times...)
			if !result.Equal(tt.expected) {
				t.Errorf("MinTime() = %v; want %v", result, tt.expected)
			}
		})
	}
}

func TestMaxTime(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)
	oneHourLater := now.Add(1 * time.Hour)

	tests := []struct {
		name     string
		times    []time.Time
		expected time.Time
	}{
		{
			name:     "empty slice",
			times:    []time.Time{},
			expected: time.Time{},
		},
		{
			name:     "single time",
			times:    []time.Time{now},
			expected: now,
		},
		{
			name:     "multiple times",
			times:    []time.Time{now, oneHourAgo, twoHoursAgo, oneHourLater},
			expected: oneHourLater,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaxTime(tt.times...)
			if !result.Equal(tt.expected) {
				t.Errorf("MaxTime() = %v; want %v", result, tt.expected)
			}
		})
	}
}
