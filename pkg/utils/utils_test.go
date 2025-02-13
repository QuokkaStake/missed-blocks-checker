package utils

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

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
	assert.Equal(t, "true", filtered[0], "Value mismatch!")
}

func TestMap(t *testing.T) {
	t.Parallel()

	array := []int{2, 4}
	filtered := Map(array, func(v int) int {
		return v * 2
	})

	assert.Len(t, filtered, 2, "Array should have 2 entries!")
	assert.Equal(t, 4, filtered[0], "Value mismatch!")
	assert.Equal(t, 8, filtered[1], "Value mismatch!")
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
		t,
		"1 day 2 hours 3 minutes 4 seconds",
		FormatDuration(duration),
		"Value mismatch!",
	)

	anotherDuration := 24 * time.Hour
	assert.Equal(
		t,
		"1 day",
		FormatDuration(anotherDuration),
		"Value mismatch!",
	)
}

func TestMakeShuffledArray(t *testing.T) {
	t.Parallel()

	array := MakeShuffledArray(10)
	assert.Len(t, array, 10, "Array should have 10 entries!")
}

func TestBoolToFloat64(t *testing.T) {
	t.Parallel()
	assert.InDelta(t, float64(1), BoolToFloat64(true), 0.001, "Value mismatch!")
	assert.InDelta(t, float64(0), BoolToFloat64(false), 0.001, "Value mismatch!")
}

func TestSplitIntoChunks(t *testing.T) {
	t.Parallel()

	array := []int{1, 2, 3, 4, 5}
	chunks := SplitIntoChunks(array, 2)

	assert.Len(t, chunks, 3, "There should be 3 chunks!")
	assert.Equal(t, []int{1, 2}, chunks[0], "Value mismatch!")
	assert.Equal(t, []int{3, 4}, chunks[1], "Value mismatch!")
	assert.Equal(t, []int{5}, chunks[2], "Value mismatch!")

	anotherArray := []int{}
	anotherChunks := SplitIntoChunks(anotherArray, 2)
	assert.Empty(t, anotherChunks, "There should be 0 chunks!")
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

func TestConvertBech32PrefixInvalidSource(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	MustConvertBech32Prefix(
		"test",
		"cosmosvaloper",
	)
}

func TestConvertBech32PrefixValid(t *testing.T) {
	t.Parallel()

	address := MustConvertBech32Prefix(
		"cosmos1xqz9pemz5e5zycaa89kys5aw6m8rhgsvtp9lt2",
		"cosmosvaloper",
	)
	assert.Equal(
		t,
		"cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e",
		address,
		"Bech addresses should not be equal!",
	)
}

func TestMaxInt64(t *testing.T) {
	t.Parallel()

	assert.Equal(t, int64(2), MaxInt64(1, 2), "Value mismatch!")
	assert.Equal(t, int64(2), MaxInt64(2, 1), "Value mismatch!")
}

func TestMinInt64(t *testing.T) {
	t.Parallel()

	assert.Equal(t, int64(1), MinInt64(1, 2), "Value mismatch!")
	assert.Equal(t, int64(1), MinInt64(2, 1), "Value mismatch!")
}

func TestSubtract(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Value string
	}

	first := []TestStruct{
		{Value: "1"},
		{Value: "2"},
		{Value: "3"},
	}

	second := []TestStruct{
		{Value: "2"},
		{Value: "4"},
	}

	result := Subtract(first, second, func(v TestStruct) any { return v.Value })
	assert.Len(t, result, 2)
	assert.Equal(t, "1", result[0].Value)
	assert.Equal(t, "3", result[1].Value)
}

func TestUnion(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Value string
	}

	first := []TestStruct{
		{Value: "1"},
		{Value: "2"},
		{Value: "3"},
	}

	second := []TestStruct{
		{Value: "2"},
		{Value: "4"},
	}

	result := Union(first, second, func(v TestStruct) any { return v.Value })
	assert.Len(t, result, 1)
	assert.Equal(t, "2", result[0].Value)
}

func TestMapToArray(t *testing.T) {
	t.Parallel()

	testMap := map[string]string{
		"test1": "1",
		"test2": "2",
		"test3": "3",
	}

	result := MapToArray(testMap)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "2")
	assert.Contains(t, result, "3")
}

func TestMustDecodeBech32Fail(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	MustDecodeBech32("invalid")
}

func TestMustDecodeBech32Ok(t *testing.T) {
	t.Parallel()

	value := MustDecodeBech32("cosmosvaloper1xqz9pemz5e5zycaa89kys5aw6m8rhgsvw4328e")
	require.Equal(t, "0600020501191b021419140204181d1d0705160410141d0e1a1b07031708100c", value)
}

func TestMustMarshallFail(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	MustJSONMarshall(make(chan bool))
}

func TestMustMarshallOk(t *testing.T) {
	t.Parallel()

	bytes := MustJSONMarshall(map[string]string{"key": "value"})
	require.JSONEq(t, "{\"key\":\"value\"}", string(bytes))
}
