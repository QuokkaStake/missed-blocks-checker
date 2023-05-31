package state

import (
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistoricalValidatorsSetValidator(t *testing.T) {
	t.Parallel()

	validators := NewHistoricalValidators()
	assert.Len(t, validators.validators, 0, "Validators should have 0 entries!")

	validators.SetValidators(10, types.HistoricalValidators{})
	assert.Len(t, validators.validators, 1, "Validators should have 1 entry!")
}

func TestHistoricalValidatorsSetAllValidators(t *testing.T) {
	t.Parallel()

	validators := NewHistoricalValidators()
	validators.SetAllValidators(types.HistoricalValidatorsMap{
		1: {},
		2: {},
	})
	assert.Len(t, validators.validators, 2, "Validators should have 2 entries!")
	assert.True(t, validators.HasValidatorsAtHeight(1), "Validators mismatch!")
	assert.True(t, validators.HasValidatorsAtHeight(2), "Validators mismatch!")
}

func TestHistoricalValidatorsGetCountSinceLatest(t *testing.T) {
	t.Parallel()

	validators := NewHistoricalValidators()
	validators.SetValidators(1, types.HistoricalValidators{})
	validators.SetValidators(3, types.HistoricalValidators{})
	validators.SetValidators(5, types.HistoricalValidators{})

	count := validators.GetCountSinceLatest(5, 5)
	assert.Equal(t, count, int64(3), "There should be 3 validators!")
}

func TestHistoricalValidatorsGetMissingSinceLatest(t *testing.T) {
	t.Parallel()

	validators := NewHistoricalValidators()
	validators.SetValidators(1, types.HistoricalValidators{})
	validators.SetValidators(3, types.HistoricalValidators{})
	validators.SetValidators(5, types.HistoricalValidators{})

	missing := validators.GetMissingSinceLatest(5, 5)
	assert.Len(t, missing, 2, "There should be 3 validators!")
	assert.Contains(t, missing, int64(2), "Validators mismatch!")
	assert.Contains(t, missing, int64(4), "Validators mismatch!")
}

func TestHistoricalValidatorsSetBlocks(t *testing.T) {
	t.Parallel()

	blocks := NewBlocks()
	blocks.SetBlocks(map[int64]*types.Block{
		1: {Height: 1},
		2: {Height: 2},
	})

	assert.True(t, blocks.HasBlockAtHeight(1), "Blocks mismatch!")
	assert.True(t, blocks.HasBlockAtHeight(2), "Blocks mismatch!")
}

func TestHistoricalValidatorsTrimBlocksBefore(t *testing.T) {
	t.Parallel()

	validators := NewHistoricalValidators()
	validators.SetValidators(1, types.HistoricalValidators{})
	validators.SetValidators(2, types.HistoricalValidators{})
	validators.SetValidators(3, types.HistoricalValidators{})

	validators.SetValidators(4, types.HistoricalValidators{})
	validators.SetValidators(5, types.HistoricalValidators{})

	validators.TrimBefore(3)

	assert.False(t, validators.HasValidatorsAtHeight(1), "Validators mismatch!")
	assert.False(t, validators.HasValidatorsAtHeight(2), "Validators mismatch!")
	assert.False(t, validators.HasValidatorsAtHeight(3), "Validators mismatch!")
	assert.True(t, validators.HasValidatorsAtHeight(4), "Validators mismatch!")
	assert.True(t, validators.HasValidatorsAtHeight(5), "Validators mismatch!")
}

func TestNewHistoricalValidatorsIsValidatorActiveAtBlock(t *testing.T) {
	t.Parallel()

	validators := NewHistoricalValidators()
	validators.SetAllValidators(types.HistoricalValidatorsMap{
		1: {"address": true},
		2: {},
	})

	validator := &types.Validator{ConsensusAddressHex: "address"}

	assert.True(t, validators.IsValidatorActiveAtBlock(validator, 1), "Validators mismatch!")
	assert.False(t, validators.IsValidatorActiveAtBlock(validator, 2), "Validators mismatch!")
	assert.False(t, validators.IsValidatorActiveAtBlock(validator, 3), "Validators mismatch!")
}
