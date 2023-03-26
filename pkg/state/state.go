package state

import (
	"main/pkg/constants"
	"main/pkg/types"
	"sync"
)

type State struct {
	Blocks          map[int64]*types.Block
	Validators      []*types.Validator
	LastBlockHeight int64
	Mutex           sync.Mutex
}

func NewState() *State {
	return &State{
		Blocks:          make(map[int64]*types.Block),
		LastBlockHeight: 0,
	}
}

func (s *State) AddBlock(block *types.Block) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.Blocks[block.Height] = block

	if block.Height > s.LastBlockHeight {
		s.LastBlockHeight = block.Height
	}
}

func (s *State) GetBlocksCountSinceLatest(expected int64) int64 {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	var expectedCount int64 = 0

	for height := s.LastBlockHeight; height > s.LastBlockHeight-expected; height-- {
		if _, ok := s.Blocks[height]; ok {
			expectedCount++
		}
	}

	return expectedCount
}

func (s *State) TrimBlocksBefore(trimHeight int64) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	for height, _ := range s.Blocks {
		if height <= trimHeight {
			delete(s.Blocks, height)
		}
	}
}

func (s *State) SetValidators(validators []*types.Validator) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	s.Validators = validators
}

func (s *State) GetValidatorMissedBlocks(validator *types.Validator, blocksToCheck int64) types.SignatureInto {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	signatureInfo := types.SignatureInto{}

	for height := s.LastBlockHeight; height > s.LastBlockHeight-blocksToCheck; height-- {
		if _, ok := s.Blocks[height]; !ok {
			continue
		}

		if s.Blocks[height].Proposer == validator.ConsensusAddress {
			signatureInfo.Proposed++
		}

		value, ok := s.Blocks[height].Signatures[validator.ConsensusAddress]

		if !ok {
			signatureInfo.NoSignature++
		} else if value != constants.ValidatorSigned && value != constants.ValidatorNilSignature {
			signatureInfo.NotSigned++
		} else {
			signatureInfo.Signed++
		}
	}

	return signatureInfo
}
