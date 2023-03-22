package utils

func Filter[T any](slice []T, f func(T) bool) []T {
	var n []T
	for _, e := range slice {
		if f(e) {
			n = append(n, e)
		}
	}
	return n
}

func Map[T, V any](slice []T, f func(T) V) []V {
	out := make([]V, len(slice))
	for index, elt := range slice {
		out[index] = f(elt)
	}
	return out
}
