package utils

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/btcsuite/btcutil/bech32"
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

func Find[T any](slice []T, f func(T) bool) (T, bool) {
	for _, elt := range slice {
		if f(elt) {
			return elt, true
		}
	}

	return *new(T), false
}

func Subtract[T any, C comparable](first, second []T, predicate func(T) C) []T {
	valuesMap := make(map[C]bool, len(second))
	for _, value := range second {
		valuesMap[predicate(value)] = true
	}

	newSlice := make([]T, 0)

	for _, value := range first {
		predicateResult := predicate(value)
		_, ok := valuesMap[predicateResult]
		if !ok {
			newSlice = append(newSlice, value)
		}
	}

	return newSlice
}

func Union[T any, C comparable](first, second []T, predicate func(T) C) []T {
	valuesMap := make(map[C]bool, len(second))
	for _, value := range second {
		valuesMap[predicate(value)] = true
	}

	newSlice := make([]T, 0)

	for _, value := range first {
		predicateResult := predicate(value)
		_, ok := valuesMap[predicateResult]
		if ok {
			newSlice = append(newSlice, value)
		}
	}

	return newSlice
}

func MapToArray[K comparable, T any](source map[K]T) []T {
	newSlice := make([]T, len(source))

	index := 0

	for _, value := range source {
		newSlice[index] = value
		index++
	}

	return newSlice
}

func SplitStringIntoChunks(msg string, maxLineLength int) []string {
	msgsByNewline := strings.Split(msg, "\n")
	outMessages := []string{}

	var sb strings.Builder

	for _, line := range msgsByNewline {
		if sb.Len()+len(line) > maxLineLength {
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
		return make([][]T, 0)
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

func MakeShuffledArray(length int) []int {
	array := make([]int, length)
	for i := range array {
		array[i] = i
	}

	rand.Shuffle(len(array), func(i, j int) {
		array[i], array[j] = array[j], array[i]
	})

	return array
}

func MustDecodeBech32(addr string) string {
	_, bech32Bytes, err := bech32.Decode(addr)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(bech32Bytes)
}

func MustConvertBech32Prefix(address, newPrefix string) string {
	_, addressRaw, err := bech32.Decode(address)
	if err != nil {
		panic(err)
	}

	encoded, _ := bech32.Encode(newPrefix, addressRaw)
	return encoded
}

func FormatDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"day", days},
		{"hour", hours},
		{"minute", minutes},
		{"second", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		case 1:
			parts = append(parts, fmt.Sprintf("%d %s", chunk.amount, chunk.singularName))
		default:
			parts = append(parts, fmt.Sprintf("%d %ss", chunk.amount, chunk.singularName))
		}
	}

	return strings.Join(parts, " ")
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}

	return b
}

func MinInt64(a, b int64) int64 {
	if a > b {
		return b
	}

	return a
}

func MustJSONMarshall(v any) []byte {
	bytes, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return bytes
}
