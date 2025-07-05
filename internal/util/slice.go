// Package util provides utility functions for working with slices
package util

import (
	"cmp"
	"slices"
	"time"

	"github.com/oleiade/reflections"
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
	slices.SortFunc(slice, func(a, b T) int {
		aVal, err := reflections.GetField(a, sortBy)
		if err != nil {
			panic(err)
		}
		bVal, err := reflections.GetField(b, sortBy)
		if err != nil {
			panic(err)
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
		panic("unsupported type for sorting slice")
	})
}
