package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntriesToSlice(t *testing.T) {
	t.Parallel()

	entries := Entries{
		"validator": Entry{
			Validator:     &Validator{Moniker: "test", Jailed: false, Status: 1},
			SignatureInfo: SignatureInto{NotSigned: 0},
		},
	}

	slice := entries.ToSlice()
	assert.NotEmpty(t, slice)
	assert.Len(t, slice, 1)
	assert.Equal(t, "test", slice[0].Validator.Moniker)
}

func TestEntriesGetActive(t *testing.T) {
	t.Parallel()

	entries := Entries{
		"firstaddr": {
			IsActive:  true,
			Validator: &Validator{Moniker: "first", OperatorAddress: "firstaddr", Status: 3},
		},
		"secondaddr": {
			IsActive:  false,
			Validator: &Validator{Moniker: "second", OperatorAddress: "secondaddr", Status: 1},
		},
	}

	activeValidators := entries.GetActive()
	assert.Len(t, activeValidators, 1)
}

func TestValidatorsGetTotalVotingPower(t *testing.T) {
	t.Parallel()

	entries := Entries{
		"firstaddr": {
			IsActive:  true,
			Validator: &Validator{Moniker: "first", OperatorAddress: "firstaddr", Status: 3, VotingPower: big.NewFloat(1)},
		},
		"secondaddr": {
			IsActive:  true,
			Validator: &Validator{Moniker: "second", OperatorAddress: "secondaddr", Status: 3, VotingPower: big.NewFloat(2)},
		},
		"thirdaddr": {
			IsActive:  false,
			Validator: &Validator{Moniker: "third", OperatorAddress: "thirdaddr", Status: 1, VotingPower: big.NewFloat(3)},
		},
	}

	totalVotingPower := entries.GetTotalVotingPower()
	assert.Equal(t, totalVotingPower, big.NewFloat(3))
}

func TestEntriesGetSoftOptOutThresholdAchievable(t *testing.T) {
	t.Parallel()

	// 3 active validators, with 80%, 15% and 5% vp, with soft-opt-out as 5%
	// top 2 should be required to sign and threshold should be 15
	entries := Entries{
		"firstaddr": {
			IsActive: true,
			Validator: &Validator{
				Moniker:         "first",
				OperatorAddress: "firstaddr",
				Status:          3,
				VotingPower:     big.NewFloat(80),
			},
		},
		"secondaddr": {
			IsActive: true,
			Validator: &Validator{
				Moniker:         "second",
				OperatorAddress: "secondaddr",
				Status:          3,
				VotingPower:     big.NewFloat(15),
			},
		},
		"thirdaddr": {
			IsActive: false,
			Validator: &Validator{
				Moniker:         "third",
				OperatorAddress: "thirdaddr",
				Status:          1,
				VotingPower:     big.NewFloat(2),
			},
		},
		"fourthaddr": {
			IsActive: true,
			Validator: &Validator{
				Moniker:         "fourth",
				OperatorAddress: "fourthaddr",
				Status:          3,
				VotingPower:     big.NewFloat(5),
			},
		},
	}

	threshold, count := entries.GetSoftOutOutThreshold(0.05)
	assert.Equal(t, big.NewFloat(15), threshold)
	assert.Equal(t, 2, count)
}

func TestEntriesGetSoftOptOutThresholdNotAchievable(t *testing.T) {
	t.Parallel()

	// 3 active validators, with 80%, 15% and 5% vp, with not achievable threshold (like -0.05)
	// it should require all active validators to be signing
	entries := Entries{
		"firstaddr": {
			IsActive: true,
			Validator: &Validator{
				Moniker:         "first",
				OperatorAddress: "firstaddr",
				Status:          3,
				VotingPower:     big.NewFloat(80),
			},
		},
		"secondaddr": {
			IsActive: true,
			Validator: &Validator{
				Moniker:         "second",
				OperatorAddress: "secondaddr",
				Status:          3,
				VotingPower:     big.NewFloat(15),
			},
		},
		"thirdaddr": {
			IsActive: false,
			Validator: &Validator{
				Moniker:         "third",
				OperatorAddress: "thirdaddr",
				Status:          1,
				VotingPower:     big.NewFloat(2),
			},
		},
		"fourthaddr": {
			IsActive: true,
			Validator: &Validator{
				Moniker:         "fourth",
				OperatorAddress: "fourthaddr",
				Status:          3,
				VotingPower:     big.NewFloat(5),
			},
		},
	}

	threshold, count := entries.GetSoftOutOutThreshold(1)
	assert.Equal(t, big.NewFloat(5), threshold)
	assert.Equal(t, 3, count)
}

func TestEntriesGetSoftOptOutThresholdEmpty(t *testing.T) {
	t.Parallel()

	entries := Entries{
		"thirdaddr": {
			IsActive: false,
			Validator: &Validator{
				Moniker:         "third",
				OperatorAddress: "thirdaddr",
				Status:          1,
				VotingPower:     big.NewFloat(2),
			},
		},
	}

	threshold, count := entries.GetSoftOutOutThreshold(0.05)
	assert.Equal(t, big.NewFloat(0), threshold)
	assert.Equal(t, 0, count)
}
