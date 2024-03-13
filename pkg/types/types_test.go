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

func TestValidatorsGetActive(t *testing.T) {
	t.Parallel()

	validators := Validators{
		{Moniker: "first", OperatorAddress: "firstaddr", Status: 3},
		{Moniker: "second", OperatorAddress: "secondaddr", Status: 1},
	}

	activeValidators := validators.GetActive()
	assert.Len(t, activeValidators, 1)
}

func TestValidatorsGetTotalVotingPower(t *testing.T) {
	t.Parallel()

	validators := Validators{
		{Moniker: "first", OperatorAddress: "firstaddr", Status: 3, VotingPower: big.NewFloat(1)},
		{Moniker: "second", OperatorAddress: "secondaddr", Status: 3, VotingPower: big.NewFloat(2)},
		{Moniker: "third", OperatorAddress: "thirdaddr", Status: 1, VotingPower: big.NewFloat(3)},
	}

	totalVotingPower := validators.GetTotalVotingPower()
	assert.Equal(t, totalVotingPower, big.NewFloat(3))
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
}

func TestValidatorsGetSoftOptOutThresholdAchievable(t *testing.T) {
	t.Parallel()

	// 3 active validators, with 80%, 15% and 5% vp, with soft-opt-out as 5%
	// top 2 should be required to sign and threshold should be 15
	validators := Validators{
		{Moniker: "first", OperatorAddress: "firstaddr", Status: 3, VotingPower: big.NewFloat(80)},
		{Moniker: "second", OperatorAddress: "secondaddr", Status: 3, VotingPower: big.NewFloat(15)},
		{Moniker: "third", OperatorAddress: "thirdaddr", Status: 1, VotingPower: big.NewFloat(2)},
		{Moniker: "fourth", OperatorAddress: "fourthaddr", Status: 3, VotingPower: big.NewFloat(5)},
	}

	threshold, count := validators.GetSoftOutOutThreshold(0.05)
	assert.Equal(t, big.NewFloat(15), threshold)
	assert.Equal(t, 2, count)
}

func TestValidatorsGetSoftOptOutThresholdNotAchievable(t *testing.T) {
	t.Parallel()

	// 3 active validators, with 80%, 15% and 5% vp, with not achievable threshold (like -0.05)
	// it should require all active validators to be signing
	validators := Validators{
		{Moniker: "first", OperatorAddress: "firstaddr", Status: 3, VotingPower: big.NewFloat(80)},
		{Moniker: "second", OperatorAddress: "secondaddr", Status: 3, VotingPower: big.NewFloat(15)},
		{Moniker: "third", OperatorAddress: "thirdaddr", Status: 1, VotingPower: big.NewFloat(2)},
		{Moniker: "fourth", OperatorAddress: "fourthaddr", Status: 3, VotingPower: big.NewFloat(5)},
	}

	threshold, count := validators.GetSoftOutOutThreshold(1)
	assert.Equal(t, big.NewFloat(5), threshold)
	assert.Equal(t, 3, count)
}

func TestValidatorsGetSoftOptOutThresholdEmpty(t *testing.T) {
	t.Parallel()

	validators := Validators{
		{Moniker: "third", OperatorAddress: "thirdaddr", Status: 1, VotingPower: big.NewFloat(2)},
	}

	threshold, count := validators.GetSoftOutOutThreshold(0.05)
	assert.Equal(t, big.NewFloat(0), threshold)
	assert.Equal(t, 0, count)
}
