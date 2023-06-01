package state

import (
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHistoricalValidatorsSetValidator(t *testing.T) {
	t.Parallel()

	state := NewState()
	assert.Len(t, state.historicalValidators.validators, 0, "Validators should have 0 entries!")

	state.AddActiveSet(10, types.HistoricalValidators{})
	assert.Len(t, state.historicalValidators.validators, 1, "Validators should have 1 entry!")
}

func TestHistoricalValidatorsSetAllValidators(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetActiveSet(types.HistoricalValidatorsMap{
		1: {},
		2: {},
	})

	assert.Len(t, state.historicalValidators.validators, 2, "Validators should have 2 entries!")
	assert.True(t, state.historicalValidators.HasValidatorsAtHeight(1), "Validators mismatch!")
	assert.True(t, state.historicalValidators.HasValidatorsAtHeight(2), "Validators mismatch!")
}

func TestHistoricalValidatorsGetCountSinceLatest(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddActiveSet(1, types.HistoricalValidators{})
	state.AddActiveSet(3, types.HistoricalValidators{})
	state.AddActiveSet(5, types.HistoricalValidators{})
	state.AddBlock(&types.Block{Height: 5})

	count := state.GetActiveSetsCountSinceLatest(5)
	assert.Equal(t, count, int64(3), "There should be 3 validators!")
}

func TestHistoricalValidatorsGetMissingSinceLatest(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddActiveSet(1, types.HistoricalValidators{})
	state.AddActiveSet(3, types.HistoricalValidators{})
	state.AddActiveSet(5, types.HistoricalValidators{})
	state.AddBlock(&types.Block{Height: 5})

	missing := state.GetMissingActiveSetsSinceLatest(5)
	assert.Len(t, missing, 2, "There should be 3 validators!")
	assert.Contains(t, missing, int64(2), "Validators mismatch!")
	assert.Contains(t, missing, int64(4), "Validators mismatch!")
}

func TestHistoricalValidatorsTrimBlocksBefore(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddActiveSet(1, types.HistoricalValidators{})
	state.AddActiveSet(2, types.HistoricalValidators{})
	state.AddActiveSet(3, types.HistoricalValidators{})

	state.AddActiveSet(4, types.HistoricalValidators{})
	state.AddActiveSet(5, types.HistoricalValidators{})

	state.TrimActiveSetsBefore(3)

	assert.False(t, state.historicalValidators.HasValidatorsAtHeight(1), "Validators mismatch!")
	assert.False(t, state.historicalValidators.HasValidatorsAtHeight(2), "Validators mismatch!")
	assert.False(t, state.historicalValidators.HasValidatorsAtHeight(3), "Validators mismatch!")
	assert.True(t, state.historicalValidators.HasValidatorsAtHeight(4), "Validators mismatch!")
	assert.True(t, state.historicalValidators.HasValidatorsAtHeight(5), "Validators mismatch!")
}

func TestNewHistoricalValidatorsIsValidatorActiveAtBlock(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetActiveSet(types.HistoricalValidatorsMap{
		1: {"address": true},
		2: {},
	})

	validator := &types.Validator{ConsensusAddressHex: "address"}

	assert.True(t, state.IsValidatorActiveAtBlock(validator, 1), "Validators mismatch!")
	assert.False(t, state.IsValidatorActiveAtBlock(validator, 2), "Validators mismatch!")
	assert.False(t, state.IsValidatorActiveAtBlock(validator, 3), "Validators mismatch!")
}
