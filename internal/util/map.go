package util

import (
	"maps"
	"slices"
)

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	return slices.Collect(maps.Keys(m))
}

func GetMapValues[K comparable, V any](m map[K]V) []V {
	return slices.Collect(maps.Values(m))
}
