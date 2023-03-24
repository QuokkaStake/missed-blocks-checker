package state

import (
	"main/pkg/types"
	"sync"
)

type State struct {
	Blocks          map[int64]*types.Block
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
	var expectedCount int64 = 0

	for height := s.LastBlockHeight; height > s.LastBlockHeight-expected; height-- {
		if _, ok := s.Blocks[height]; ok {
			expectedCount++
		}
	}

	return expectedCount
}

func (s *State) TrimBlocksBefore(trimHeight int64) {
	for height, _ := range s.Blocks {
		if height <= trimHeight {
			delete(s.Blocks, height)
		}
	}
}
