package types

import (
	"main/pkg/utils"
	"math/big"
	"sort"
)

type Entry struct {
	IsActive      bool
	NeedsToSign   bool
	Validator     *Validator
	SignatureInfo SignatureInto

	VotingPowerPercent           float64
	CumulativeVotingPowerPercent float64
	Rank                         int
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

func (e Entries) GetTotalVotingPower() *big.Float {
	sum := big.NewFloat(0)

	for _, entry := range e {
		if entry.IsActive {
			sum.Add(sum, entry.Validator.VotingPower)
		}
	}

	return sum
}

func (e Entries) GetSoftOutOutThreshold(softOptOut float64) (*big.Float, int) {
	sortedEntries := e.GetActive()

	if len(sortedEntries) == 0 {
		return big.NewFloat(0), 0
	}

	// sorting validators by voting power ascending
	sort.Slice(sortedEntries, func(first, second int) bool {
		return sortedEntries[first].Validator.VotingPower.Cmp(sortedEntries[second].Validator.VotingPower) < 0
	})

	totalVP := e.GetTotalVotingPower()
	threshold := big.NewFloat(0)

	for index, validator := range sortedEntries {
		threshold = big.NewFloat(0).Add(threshold, validator.Validator.VotingPower)
		thresholdPercent := big.NewFloat(0).Quo(threshold, totalVP)

		if thresholdPercent.Cmp(big.NewFloat(softOptOut)) > 0 {
			return validator.Validator.VotingPower, index + 1
		}
	}

	// should've never reached here
	return sortedEntries[0].Validator.VotingPower, len(sortedEntries)
}

func (e Entries) SetVotingPowerPercent() {
	totalVP := e.GetTotalVotingPower()

	activeAndSortedEntries := e.GetActive()

	// sorting by voting power desc
	sort.Slice(activeAndSortedEntries, func(first, second int) bool {
		return activeAndSortedEntries[first].Validator.VotingPower.Cmp(activeAndSortedEntries[second].Validator.VotingPower) > 0
	})

	var cumulativeVotingPowerPercent float64 = 0
	for index, sortedEntry := range activeAndSortedEntries {
		percent, _ := new(big.Float).Quo(sortedEntry.Validator.VotingPower, totalVP).Float64()

		entry := e[sortedEntry.Validator.OperatorAddress]

		entry.VotingPowerPercent = percent
		entry.Rank = index + 1

		cumulativeVotingPowerPercent += percent
		entry.CumulativeVotingPowerPercent = cumulativeVotingPowerPercent
	}
}
