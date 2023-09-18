package utils

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func StringOfRandomLength(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func TestFilter(t *testing.T) {
	t.Parallel()

	array := []string{"true", "false"}
	filtered := Filter(array, func(s string) bool {
		return s == "true"
	})

	assert.Len(t, filtered, 1, "Array should have 1 entry!")
	assert.Equal(t, filtered[0], "true", "Value mismatch!")
}

func TestMap(t *testing.T) {
	t.Parallel()

	array := []int{2, 4}
	filtered := Map(array, func(v int) int {
		return v * 2
	})

	assert.Len(t, filtered, 2, "Array should have 2 entries!")
	assert.Equal(t, filtered[0], 4, "Value mismatch!")
	assert.Equal(t, filtered[1], 8, "Value mismatch!")
}

func TestContains(t *testing.T) {
	t.Parallel()

	array := []string{"true", "false"}

	assert.True(t, Contains(array, "true"), "Array should contain value!")
	assert.True(t, Contains(array, "false"), "Array should contain value!")
	assert.False(t, Contains(array, "not"), "Array should not contain value!")
}

func TestContainsFound(t *testing.T) {
	t.Parallel()

	array := []string{"true", "false"}

	value, found := Find(array, func(s string) bool {
		return s == "true"
	})
	assert.NotNil(t, value, "Value should be presented!")
	assert.True(t, found, "Value should be found!")
}

func TestContainsNotFound(t *testing.T) {
	t.Parallel()

	array := []string{"true", "false"}

	_, found := Find(array, func(s string) bool {
		return s == "test"
	})
	assert.False(t, found, "Value should not be found!")
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	duration := 24*time.Hour + 2*time.Hour + 3*time.Minute + 4*time.Second
	assert.Equal(
		t, FormatDuration(duration),
		"1 day 2 hours 3 minutes 4 seconds",
		"Value mismatch!",
	)

	anotherDuration := 24 * time.Hour
	assert.Equal(
		t, FormatDuration(anotherDuration),
		"1 day",
		"Value mismatch!",
	)
}

func TestMakeShuffledArray(t *testing.T) {
	t.Parallel()

	array := MakeShuffledArray(10)
	assert.Len(t, array, 10, "Array should have 10 entries!")
}

func TestCompareTwoBech32FirstInvalid(t *testing.T) {
	t.Parallel()

	_, err := CompareTwoBech32("test", "cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2")
	assert.NotNil(t, err, "Error should be present!")
}

func TestCompareTwoBech32SecondInvalid(t *testing.T) {
	t.Parallel()

	_, err := CompareTwoBech32("cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2", "test")
	assert.NotNil(t, err, "Error should be present!")
}

func TestCompareTwoBech32SecondEqual(t *testing.T) {
	t.Parallel()

	equal, err := CompareTwoBech32(
		"cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
		"cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
	)
	assert.Nil(t, err, "Error should not be present!")
	assert.True(t, equal, "Bech addresses should be equal!")
}

func TestCompareTwoBech32SecondNotEqual(t *testing.T) {
	t.Parallel()

	equal, err := CompareTwoBech32(
		"cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
		"cosmos1c4k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q",
	)
	assert.Nil(t, err, "Error should not be present!")
	assert.False(t, equal, "Bech addresses should not be equal!")
}

func TestCompareBoolToFloat64(t *testing.T) {
	t.Parallel()
	assert.Equal(t, BoolToFloat64(true), float64(1), "Value mismatch!")
	assert.Equal(t, BoolToFloat64(false), float64(0), "Value mismatch!")
}

func TestSplitIntoChunks(t *testing.T) {
	t.Parallel()

	array := []int{1, 2, 3, 4, 5}
	chunks := SplitIntoChunks(array, 2)

	assert.Len(t, chunks, 3, "There should be 3 chunks!")
	assert.Equal(t, chunks[0], []int{1, 2}, "Value mismatch!")
	assert.Equal(t, chunks[1], []int{3, 4}, "Value mismatch!")
	assert.Equal(t, chunks[2], []int{5}, "Value mismatch!")

	anotherArray := []int{}
	anotherChunks := SplitIntoChunks(anotherArray, 2)
	assert.Len(t, anotherChunks, 0, "There should be 0 chunks!")
}

func TestSplitStringIntoChunksLessThanOneChunk(t *testing.T) {
	t.Parallel()

	str := StringOfRandomLength(10)
	chunks := SplitStringIntoChunks(str, 20)
	assert.Len(t, chunks, 1, "There should be 1 chunk!")
}

func TestSplitStringIntoChunksExactlyOneChunk(t *testing.T) {
	t.Parallel()

	str := StringOfRandomLength(10)
	chunks := SplitStringIntoChunks(str, 10)

	assert.Len(t, chunks, 1, "There should be 1 chunk!")
}

func TestSplitStringIntoChunksMoreChunks(t *testing.T) {
	t.Parallel()

	str := "aaaa\nbbbb\ncccc\ndddd\neeeee\n"
	chunks := SplitStringIntoChunks(str, 10)
	assert.Len(t, chunks, 3, "There should be 3 chunks!")
}

func TestConvertBech32PrefixInvalid(t *testing.T) {
	t.Parallel()

	_, err := ConvertBech32Prefix(
		"test",
		"cosmosvaloper",
	)
	assert.NotNil(t, err, "Error should be present!")
}

func TestConvertBech32PrefixValid(t *testing.T) {
	t.Parallel()

	address, err := ConvertBech32Prefix(
		"cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
		"cosmosvaloper",
	)
	assert.Nil(t, err, "Error should not be present!")
	assert.Equal(
		t,
		address,
		"cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
		"Bech addresses should not be equal!",
	)
}

func TestMaxInt64(t *testing.T) {
	t.Parallel()

	assert.Equal(t, MaxInt64(1, 2), int64(2), "Value mismatch!")
	assert.Equal(t, MaxInt64(2, 1), int64(2), "Value mismatch!")
}

func TestMinInt64(t *testing.T) {
	t.Parallel()

	assert.Equal(t, MinInt64(1, 2), int64(1), "Value mismatch!")
	assert.Equal(t, MinInt64(2, 1), int64(1), "Value mismatch!")
}
