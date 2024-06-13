package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorsToMap(t *testing.T) {
	t.Parallel()

	validators := Validators{
		{Moniker: "first", OperatorAddress: "firstaddr"},
		{Moniker: "second", OperatorAddress: "secondaddr"},
	}

	validatorsMap := validators.ToMap()
	assert.Len(t, validatorsMap, 2, "Map should have 2 entries!")
	assert.Equal(t, "first", validatorsMap["firstaddr"].Moniker, "Validator mismatch!")
	assert.Equal(t, "second", validatorsMap["secondaddr"].Moniker, "Validator mismatch!")
}

func TestValidatorsToSlice(t *testing.T) {
	t.Parallel()

	validatorsMap := ValidatorsMap{
		"firstaddr":  {Moniker: "first", OperatorAddress: "firstaddr"},
		"secondaddr": {Moniker: "second", OperatorAddress: "secondaddr"},
	}

	validators := validatorsMap.ToSlice()
	assert.Len(t, validators, 2, "Slice should have 2 entries!")

	monikers := []string{
		validators[0].Moniker,
		validators[1].Moniker,
	}

	assert.Contains(t, monikers, "first", "Validator mismatch!")
	assert.Contains(t, monikers, "second", "Validator mismatch!")
}

func TestValidatorsSetTotalVotingPower(t *testing.T) {
	t.Parallel()

	validators := Validators{
		{Moniker: "first", OperatorAddress: "firstaddr", Status: 3, VotingPower: big.NewFloat(1)},
		{Moniker: "second", OperatorAddress: "secondaddr", Status: 3, VotingPower: big.NewFloat(3)},
		{Moniker: "third", OperatorAddress: "thirdaddr", Status: 1, VotingPower: big.NewFloat(2)},
	}

	validators.SetVotingPowerPercent()
	assert.Len(t, validators, 3)
	assert.InDelta(t, 0.25, validators[0].VotingPowerPercent, 0.001)
	assert.InDelta(t, 0.75, validators[1].VotingPowerPercent, 0.001)
	assert.Equal(t, 2, validators[0].Rank)
	assert.Equal(t, 1, validators[1].Rank)
	assert.InDelta(t, float64(1), validators[0].CumulativeVotingPowerPercent, 0.001)
	assert.InDelta(t, 0.75, validators[1].CumulativeVotingPowerPercent, 0.001)
}
