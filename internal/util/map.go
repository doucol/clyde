package util

import (
	"maps"
	"slices"
	"strings"
)

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	return slices.Collect(maps.Keys(m))
}

func GetMapValues[K comparable, V any](m map[K]V) []V {
	return slices.Collect(maps.Values(m))
}

func JoinMapKeys(m map[string]any, sep string) string {
	return strings.Join(GetMapKeys(m), sep)
}

func JoinMapValues(m map[any]string, sep string) string {
	return strings.Join(GetMapValues(m), sep)
}
