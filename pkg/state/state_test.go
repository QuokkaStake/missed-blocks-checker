package state

import (
	"github.com/stretchr/testify/assert"
	"main/pkg/types"
	"testing"
)

func TestStateGetAddAndLatestBlock(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 10})
	assert.Equal(t, state.GetLastBlockHeight(), int64(10), "Height mismatch!")
}

func TestStateSetAndGetValidators(t *testing.T) {
	t.Parallel()

	state := NewState()

	validators := types.ValidatorsMap{
		"address": &types.Validator{OperatorAddress: "address", Moniker: "moniker"},
	}

	state.SetValidators(validators)
	validatorsFromState := state.GetValidators()

	assert.Len(t, validatorsFromState, 1, "Length mismatch!")
	assert.Equal(t, validatorsFromState["address"].Moniker, "moniker", "Validator mismatch!")

	validatorFromState, found := state.GetValidator("address")

	assert.True(t, found, 1, "Validator should be present!")
	assert.Equal(t, validatorFromState.Moniker, "moniker", "Validator mismatch!")
}
