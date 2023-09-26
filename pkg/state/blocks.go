package state

import (
	"main/pkg/types"
	"main/pkg/utils"
	"sync"
)

type Blocks struct {
	mutex      sync.RWMutex
	blocks     types.BlocksMap
	lastHeight int64
}

func NewBlocks() *Blocks {
	return &Blocks{
		blocks:     make(types.BlocksMap),
		lastHeight: 0,
	}
}

func (b *Blocks) AddBlock(block *types.Block) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.blocks[block.Height] = block

	if block.Height > b.lastHeight {
		b.lastHeight = block.Height
	}
}

func (b *Blocks) HasBlockAtHeight(height int64) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	_, ok := b.blocks[height]
	return ok
}

func (b *Blocks) TrimBefore(trimHeight int64) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for height := range b.blocks {
		if height <= trimHeight {
			delete(b.blocks, height)
		}
	}
}

func (b *Blocks) SetBlocks(blocks map[int64]*types.Block) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.blocks = blocks

	var lastHeight int64 = 0

	for height := range b.blocks {
		if height > lastHeight {
			lastHeight = height
		}
	}

	b.lastHeight = lastHeight
}

func (b *Blocks) GetBlock(height int64) (*types.Block, bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	block, ok := b.blocks[height]
	return block, ok
}

func (b *Blocks) GetLatestBlock() *types.Block {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.blocks[b.lastHeight]
}

func (b *Blocks) GetEarliestBlock() *types.Block {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	earliestHeight := b.lastHeight

	for height := range b.blocks {
		if height < earliestHeight {
			earliestHeight = height
		}
	}

	return b.blocks[earliestHeight]
}

func (b *Blocks) GetCountSinceLatest(expected int64) int64 {
	var expectedCount int64 = 0

	for height := b.lastHeight; height > b.lastHeight-expected; height-- {
		if b.HasBlockAtHeight(height) {
			expectedCount++
		}
	}

	return expectedCount
}

func (b *Blocks) GetMissingSinceLatest(expected int64) []int64 {
	var missed []int64

	earliestBlock := utils.MaxInt64(b.lastHeight-expected, 0)

	for height := b.lastHeight; height > earliestBlock; height-- {
		if !b.HasBlockAtHeight(height) {
			missed = append(missed, height)
		}
	}

	return missed
}
