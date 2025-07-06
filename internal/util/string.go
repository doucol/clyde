package util

import (
	"slices"
	"strings"
)

func DedupDelimitedStrings(sep string, source ...string) []string {
	set := make(map[string]any)
	setval := func(val string) {
		v := strings.TrimSpace(val)
		if v != "" {
			set[v] = struct{}{}
		}
	}
	for _, str := range source {
		sl := strings.SplitSeq(str, sep)
		for item := range sl {
			setval(item)
		}
	}
	sorted := GetMapKeys(set)
	slices.Sort(sorted)
	return sorted
}

func NormalizeLabels(sources ...string) string {
	return strings.Join(DedupDelimitedStrings("|", sources...), " | ")
}
