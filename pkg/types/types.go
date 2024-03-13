package types

import (
	"math/big"
	"sort"
)

type WebsocketEmittable interface {
	Hash() string
}

type Link struct {
	Href string
	Text string
}

type BlocksMap map[int64]*Block
type Validators []*Validator
type ValidatorsMap map[string]*Validator

func (validators Validators) ToMap() ValidatorsMap {
	validatorsMap := make(ValidatorsMap, len(validators))

	for _, validator := range validators {
		validatorsMap[validator.OperatorAddress] = validator
	}

	return validatorsMap
}

func (validators Validators) GetActive() Validators {
	activeValidators := make(Validators, 0)
	for _, validator := range validators {
		if validator.Active() {
			activeValidators = append(activeValidators, validator)
		}
	}

	return activeValidators
}

func (validators Validators) GetTotalVotingPower() *big.Float {
	sum := big.NewFloat(0)

	for _, validator := range validators {
		if validator.Active() {
			sum.Add(sum, validator.VotingPower)
		}
	}

	return sum
}

func (validators Validators) SetVotingPowerPercent() {
	totalVP := validators.GetTotalVotingPower()

	activeAndSortedValidators := validators.GetActive()

	// sorting by voting power desc
	sort.Slice(activeAndSortedValidators, func(first, second int) bool {
		return activeAndSortedValidators[first].VotingPower.Cmp(activeAndSortedValidators[second].VotingPower) > 0
	})

	var cumulativeVotingPowerPercent float64 = 0
	for index, validator := range activeAndSortedValidators {
		percent, _ := new(big.Float).Quo(validator.VotingPower, totalVP).Float64()
		validator.VotingPowerPercent = percent
		validator.Rank = index + 1

		cumulativeVotingPowerPercent += percent
		validator.CumulativeVotingPowerPercent = cumulativeVotingPowerPercent
	}
}

func (validators Validators) GetSoftOutOutThreshold(softOptOut float64) (*big.Float, int) {
	sortedValidators := validators.GetActive()

	if len(sortedValidators) == 0 {
		return big.NewFloat(0), 0
	}

	// sorting validators by voting power ascending
	sort.Slice(sortedValidators, func(first, second int) bool {
		return sortedValidators[first].VotingPower.Cmp(sortedValidators[second].VotingPower) < 0
	})

	totalVP := validators.GetTotalVotingPower()
	threshold := big.NewFloat(0)

	for index, validator := range sortedValidators {
		threshold = big.NewFloat(0).Add(threshold, validator.VotingPower)
		thresholdPercent := big.NewFloat(0).Quo(threshold, totalVP)

		if thresholdPercent.Cmp(big.NewFloat(softOptOut)) > 0 {
			return validator.VotingPower, index + 1
		}
	}

	// should've never reached here
	return sortedValidators[0].VotingPower, len(sortedValidators)
}

func (validatorsMap ValidatorsMap) ToSlice() Validators {
	validators := make(Validators, len(validatorsMap))
	index := 0

	for _, validator := range validatorsMap {
		validators[index] = validator
		index++
	}

	return validators
}
