package utils

import (
	"strings"
)

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

func Contains[T comparable](slice []T, elt T) bool {
	for _, innerElt := range slice {
		if elt == innerElt {
			return true
		}
	}

	return false
}

func Find[T any](slice []*T, f func(*T) bool) (*T, bool) {
	for _, elt := range slice {
		if f(elt) {
			return elt, true
		}
	}

	return nil, false
}

func SplitStringIntoChunks(msg string, maxLineLength int) []string {
	msgsByNewline := strings.Split(msg, "\n")
	outMessages := []string{}

	var sb strings.Builder

	for _, line := range msgsByNewline {
		if sb.Len()+len(line) >= maxLineLength {
			outMessages = append(outMessages, sb.String())
			sb.Reset()
		}

		sb.WriteString(line + "\n")
	}

	outMessages = append(outMessages, sb.String())
	return outMessages
}

func BoolToFloat64(value bool) float64 {
	if value {
		return 1
	}

	return 0
}

func SplitIntoChunks[T any](items []T, chunkSize int) [][]T {
	if len(items) == 0 {
		return nil
	}

	divided := make([][]T, (len(items)+chunkSize-1)/chunkSize)
	prev := 0
	i := 0
	till := len(items) - chunkSize
	for prev < till {
		next := prev + chunkSize
		divided[i] = items[prev:next]
		prev = next
		i++
	}
	divided[i] = items[prev:]
	return divided
}
