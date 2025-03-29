package util

// The accept func should return true to keep the element
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
