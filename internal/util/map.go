package util

import "strings"

func GetMapKeys[K comparable, V any](m map[K]V) []K {
	i := 0
	keys := make([]K, len(m))
	for k := range m {
		keys[i] = k
		i++
	}
	return keys
}

func JoinMapKeys(m map[string]any, sep string) string {
	keys := GetMapKeys(m)
	return strings.Join(keys, sep)
}
