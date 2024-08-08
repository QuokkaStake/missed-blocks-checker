package types

import (
	"main/pkg/utils"
	"sort"

	"cosmossdk.io/math"
)

type Entry struct {
	IsActive      bool
	NeedsToSign   bool
	Validator     *Validator
	SignatureInfo SignatureInto
}

type Entries map[string]*Entry

func (e Entries) ToSlice() []*Entry {
	entries := make([]*Entry, len(e))

	index := 0
	for _, entry := range e {
		entries[index] = entry
		index++
	}

	return entries
}

func (e Entries) ByValidatorAddresses(addresses []string) []*Entry {
	entries := make([]*Entry, 0)

	for _, entry := range e {
		if utils.Contains(addresses, entry.Validator.OperatorAddress) {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (e Entries) GetActive() []*Entry {
	activeValidators := make([]*Entry, 0)
	for _, entry := range e {
		if entry.IsActive {
			activeValidators = append(activeValidators, entry)
		}
	}

	return activeValidators
}

func (e Entries) GetTotalVotingPower() math.LegacyDec {
	sum := math.LegacyZeroDec()

	for _, entry := range e {
		if entry.IsActive {
			sum = sum.Add(entry.Validator.VotingPower)
		}
	}

	return sum
}

func (e Entries) SetVotingPowerPercent() {
	totalVP := e.GetTotalVotingPower()

	activeAndSortedEntries := e.GetActive()

	// sorting by voting power desc
	sort.Slice(activeAndSortedEntries, func(first, second int) bool {
		return activeAndSortedEntries[first].Validator.VotingPower.GT(activeAndSortedEntries[second].Validator.VotingPower)
	})

	var cumulativeVotingPowerPercent float64 = 0
	for index, sortedEntry := range activeAndSortedEntries {
		percent := sortedEntry.Validator.VotingPower.Quo(totalVP).MustFloat64()

		entry := e[sortedEntry.Validator.OperatorAddress]

		entry.Validator.VotingPowerPercent = percent
		entry.Validator.Rank = index + 1

		cumulativeVotingPowerPercent += percent
		entry.Validator.CumulativeVotingPowerPercent = cumulativeVotingPowerPercent
	}
}
