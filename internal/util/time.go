package util

import "time"

func MinTime(times ...time.Time) time.Time {
	if len(times) == 0 {
		return time.Time{} // Return zero value if slice is empty
	}

	min := times[0]
	for _, t := range times[1:] {
		if t.Before(min) {
			min = t
		}
	}
	return min
}

func MaxTime(times ...time.Time) time.Time {
	if len(times) == 0 {
		return time.Time{} // Return zero value if slice is empty
	}

	max := times[0]
	for _, t := range times[1:] {
		if t.After(max) {
			max = t
		}
	}
	return max
}
