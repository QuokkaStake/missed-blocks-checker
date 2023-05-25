package state

import (
	"main/pkg/types"
	"sync"
)

type HistoricalValidators struct {
	mutex      sync.RWMutex
	validators types.HistoricalValidatorsMap
	lastHeight int64
}

func NewHistoricalValidators() *HistoricalValidators {
	return &HistoricalValidators{
		validators: make(types.HistoricalValidatorsMap),
		lastHeight: 0,
	}
}

func (h *HistoricalValidators) SetAllValidators(validators types.HistoricalValidatorsMap) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.validators = validators

	for height := range h.validators {
		if height > h.lastHeight {
			h.lastHeight = height
		}
	}
}

func (h *HistoricalValidators) SetValidators(height int64, validators types.HistoricalValidators) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.validators[height] = validators

	if height > h.lastHeight {
		h.lastHeight = height
	}
}

func (h *HistoricalValidators) TrimBefore(trimHeight int64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for height := range h.validators {
		if height <= trimHeight {
			delete(h.validators, height)
		}
	}
}

func (h *HistoricalValidators) IsValidatorActiveAtBlock(validator *types.Validator, height int64) bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.validators[height]; !ok {
		return false
	}

	_, ok := h.validators[height][validator.ConsensusAddressHex]
	return ok
}

func (h *HistoricalValidators) GetCountSinceLatest(expected int64, lastHeight int64) int64 {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var expectedCount int64 = 0

	for height := lastHeight; height > lastHeight-expected; height-- {
		if _, ok := h.validators[height]; ok {
			expectedCount++
		}
	}

	return expectedCount
}

func (h *HistoricalValidators) GetMissingSinceLatest(expected int64, lastHeight int64) []int64 {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	var missing []int64

	for height := lastHeight; height > lastHeight-expected; height-- {
		if _, ok := h.validators[height]; !ok {
			missing = append(missing, height)
		}
	}

	return missing
}
