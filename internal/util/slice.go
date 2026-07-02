// Package util provides utility functions for working with slices
package util

import (
	"cmp"
	"slices"
	"time"

	"github.com/oleiade/reflections"
	"github.com/sirupsen/logrus"
)

func FilterSlice[T any](slice []T, accept func(T) bool) []T {
	i := 0
	for _, elem := range slice {
		if accept(elem) {
			slice[i] = elem
			i++
		}
	}
	return slice[:i]
}

func SortSlice[T any](slice []T, sortBy string, ascending bool) {
	if len(slice) == 0 {
		return
	}
	// Validate the sort field once up front; if it is missing or of an
	// unsupported type, leave the slice unsorted rather than panicking inside
	// the comparator (which runs on the UI path).
	if v, err := reflections.GetField(slice[0], sortBy); err != nil {
		logrus.WithError(err).Errorf("cannot sort by field %q; leaving order unchanged", sortBy)
		return
	} else {
		switch v.(type) {
		case string, int, int64, uint64, float64, time.Time, int32:
		default:
			logrus.Errorf("unsupported type for sorting by field %q; leaving order unchanged", sortBy)
			return
		}
	}
	slices.SortFunc(slice, func(a, b T) int {
		aVal, err := reflections.GetField(a, sortBy)
		if err != nil {
			return 0
		}
		bVal, err := reflections.GetField(b, sortBy)
		if err != nil {
			return 0
		}
		switch av := aVal.(type) {
		case string:
			if ascending {
				return cmp.Compare(av, bVal.(string))
			} else {
				return cmp.Compare(bVal.(string), av)
			}
		case int:
			if ascending {
				return cmp.Compare(av, bVal.(int))
			} else {
				return cmp.Compare(bVal.(int), av)
			}
		case int64:
			if ascending {
				return cmp.Compare(av, bVal.(int64))
			} else {
				return cmp.Compare(bVal.(int64), av)
			}
		case uint64:
			if ascending {
				return cmp.Compare(av, bVal.(uint64))
			} else {
				return cmp.Compare(bVal.(uint64), av)
			}
		case float64:
			if ascending {
				return cmp.Compare(av, bVal.(float64))
			} else {
				return cmp.Compare(bVal.(float64), av)
			}
		case time.Time:
			if ascending {
				return av.Compare(bVal.(time.Time))
			} else {
				return bVal.(time.Time).Compare(av)
			}
		case int32:
			if ascending {
				return cmp.Compare(av, bVal.(int32))
			} else {
				return cmp.Compare(bVal.(int32), av)
			}
		}
		// Unreachable: the field type was validated above.
		return 0
	})
}
