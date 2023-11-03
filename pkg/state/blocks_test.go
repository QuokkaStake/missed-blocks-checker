package state

import (
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlocksAddBlock(t *testing.T) {
	t.Parallel()

	state := NewState()
	assert.Empty(t, state.blocks.blocks, "Blocks should have 0 entries!")

	state.AddBlock(&types.Block{Height: 10})
	assert.NotEmpty(t, state.blocks.blocks, "Blocks should have 1 entry!")
}

func TestBlocksGetLatestBlock(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 20})
	state.AddBlock(&types.Block{Height: 10})
	state.AddBlock(&types.Block{Height: 30})

	latest := state.blocks.GetLatestBlock()

	assert.Equal(t, int64(30), latest.Height, "Height mismatch!")
}

func TestBlocksGetEarliestBlock(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 20})
	state.AddBlock(&types.Block{Height: 10})
	state.AddBlock(&types.Block{Height: 30})

	latest := state.GetEarliestBlock()

	assert.Equal(t, int64(10), latest.Height, "Height mismatch!")
}

func TestBlocksGetBlock(t *testing.T) {
	t.Parallel()

	blocks := NewBlocks()
	blocks.AddBlock(&types.Block{Height: 10})

	block, found := blocks.GetBlock(10)
	assert.NotNil(t, block, "Block should be found!")
	assert.True(t, found, "Block should be found!")

	anotherBlock, anotherFound := blocks.GetBlock(20)
	assert.Nil(t, anotherBlock, "Block should not be found!")
	assert.False(t, anotherFound, "Block should not be found!")
}

func TestBlocksHasBlock(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 10})

	found := state.HasBlockAtHeight(10)
	assert.True(t, found, "Block should be found!")

	anotherFound := state.HasBlockAtHeight(20)
	assert.False(t, anotherFound, "Block should not be found!")
}

func TestBlocksGetCountSinceLatest(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 1})
	state.AddBlock(&types.Block{Height: 3})
	state.AddBlock(&types.Block{Height: 5})

	count := state.GetBlocksCountSinceLatest(5)
	assert.Equal(t, int64(3), count, "There should be 3 blocks!")
}

func TestBlocksGetMissingSinceLatest(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 1})
	state.AddBlock(&types.Block{Height: 3})
	state.AddBlock(&types.Block{Height: 5})

	missing := state.GetMissingBlocksSinceLatest(5)

	assert.Len(t, missing, 2, "There should be 3 blocks!")
	assert.Contains(t, missing, int64(2), "Blocks mismatch!")
	assert.Contains(t, missing, int64(4), "Blocks mismatch!")
}

func TestBlocksGetMissingSinceLatestNotEnoughBlocks(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 1})
	state.AddBlock(&types.Block{Height: 3})

	missing := state.GetMissingBlocksSinceLatest(5)
	assert.Len(t, missing, 1, "There should be 2 blocks!")
	assert.Contains(t, missing, int64(2), "Blocks mismatch!")
}

func TestBlocksSetBlocks(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.SetBlocks(map[int64]*types.Block{
		1: {Height: 1},
		2: {Height: 2},
	})

	assert.True(t, state.HasBlockAtHeight(1), "Blocks mismatch!")
	assert.True(t, state.HasBlockAtHeight(2), "Blocks mismatch!")
}

func TestBlocksTrimBlocksBefore(t *testing.T) {
	t.Parallel()

	state := NewState()
	state.AddBlock(&types.Block{Height: 1})
	state.AddBlock(&types.Block{Height: 2})
	state.AddBlock(&types.Block{Height: 3})
	state.AddBlock(&types.Block{Height: 4})
	state.AddBlock(&types.Block{Height: 5})

	state.TrimBlocksBefore(3)

	assert.False(t, state.HasBlockAtHeight(1), "Blocks mismatch!")
	assert.False(t, state.HasBlockAtHeight(2), "Blocks mismatch!")
	assert.False(t, state.HasBlockAtHeight(3), "Blocks mismatch!")
	assert.True(t, state.HasBlockAtHeight(4), "Blocks mismatch!")
	assert.True(t, state.HasBlockAtHeight(5), "Blocks mismatch!")
}
