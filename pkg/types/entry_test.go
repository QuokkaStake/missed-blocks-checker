package types

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntriesToSlice(t *testing.T) {
	t.Parallel()

	entries := Entries{
		"validator": &Entry{
			Validator:     &Validator{Moniker: "test", Jailed: false},
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
			Validator: &Validator{Moniker: "first", OperatorAddress: "firstaddr"},
		},
		"secondaddr": {
			IsActive:  false,
			Validator: &Validator{Moniker: "second", OperatorAddress: "secondaddr"},
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
			Validator: &Validator{Moniker: "first", OperatorAddress: "firstaddr", VotingPower: big.NewFloat(1)},
		},
		"secondaddr": {
			IsActive:  true,
			Validator: &Validator{Moniker: "second", OperatorAddress: "secondaddr", VotingPower: big.NewFloat(2)},
		},
		"thirdaddr": {
			IsActive:  false,
			Validator: &Validator{Moniker: "third", OperatorAddress: "thirdaddr", VotingPower: big.NewFloat(3)},
		},
	}

	totalVotingPower := entries.GetTotalVotingPower()
	assert.Equal(t, totalVotingPower, big.NewFloat(3))
}

func TestEntriesSetTotalVotingPower(t *testing.T) {
	t.Parallel()

	entries := Entries{
		"firstaddr": {
			IsActive: true,
			Validator: &Validator{
				Moniker:         "first",
				OperatorAddress: "firstaddr",
				VotingPower:     big.NewFloat(1),
			},
		},
		"secondaddr": {
			IsActive: true,
			Validator: &Validator{
				Moniker:         "second",
				OperatorAddress: "secondaddr",
				VotingPower:     big.NewFloat(3),
			},
		},
		"thirdaddr": {
			IsActive: false,
			Validator: &Validator{
				Moniker:         "third",
				OperatorAddress: "thirdaddr",
				VotingPower:     big.NewFloat(2),
			},
		},
	}

	entries.SetVotingPowerPercent()
	assert.Len(t, entries, 3)
	assert.InDelta(t, 0.25, entries["firstaddr"].VotingPowerPercent, 0.001)
	assert.InDelta(t, 0.75, entries["secondaddr"].VotingPowerPercent, 0.001)
	assert.Equal(t, 2, entries["firstaddr"].Rank)
	assert.Equal(t, 1, entries["secondaddr"].Rank)
	assert.InDelta(t, float64(1), entries["firstaddr"].CumulativeVotingPowerPercent, 0.001)
	assert.InDelta(t, 0.75, entries["secondaddr"].CumulativeVotingPowerPercent, 0.001)
}

func TestEntriesGetByValidatorAddresses(t *testing.T) {
	t.Parallel()

	entries := Entries{
		"firstaddr":  {Validator: &Validator{OperatorAddress: "firstaddr"}},
		"secondaddr": {Validator: &Validator{OperatorAddress: "secondaddr"}},
		"thirdaddr":  {Validator: &Validator{OperatorAddress: "thirdaddr"}},
	}

	filteredEntries := entries.ByValidatorAddresses([]string{"firstaddr", "secondaddr"})
	assert.Len(t, filteredEntries, 2)
}
