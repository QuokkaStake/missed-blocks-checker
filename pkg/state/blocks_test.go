package state

import (
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlocksAddBlock(t *testing.T) {
	t.Parallel()

	blocks := NewBlocks()
	assert.Len(t, blocks.blocks, 0, "Blocks should have 0 entries!")

	blocks.AddBlock(&types.Block{Height: 10})
	assert.Len(t, blocks.blocks, 1, "Blocks should have 1 entry!")
}

func TestBlocksGetLatestBlock(t *testing.T) {
	t.Parallel()

	blocks := NewBlocks()
	blocks.AddBlock(&types.Block{Height: 20})
	blocks.AddBlock(&types.Block{Height: 10})
	blocks.AddBlock(&types.Block{Height: 30})

	latest := blocks.GetLatestBlock()

	assert.Equal(t, latest.Height, int64(30), "Height mismatch!")
}

func TestBlocksGetEarliestBlock(t *testing.T) {
	t.Parallel()

	blocks := NewBlocks()
	blocks.AddBlock(&types.Block{Height: 20})
	blocks.AddBlock(&types.Block{Height: 10})
	blocks.AddBlock(&types.Block{Height: 30})

	latest := blocks.GetEarliestBlock()

	assert.Equal(t, latest.Height, int64(10), "Height mismatch!")
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

	blocks := NewBlocks()
	blocks.AddBlock(&types.Block{Height: 10})

	found := blocks.HasBlockAtHeight(10)
	assert.True(t, found, "Block should be found!")

	anotherFound := blocks.HasBlockAtHeight(20)
	assert.False(t, anotherFound, "Block should not be found!")
}

func TestBlocksGetCountSinceLatest(t *testing.T) {
	t.Parallel()

	blocks := NewBlocks()
	blocks.AddBlock(&types.Block{Height: 1})
	blocks.AddBlock(&types.Block{Height: 3})
	blocks.AddBlock(&types.Block{Height: 5})

	count := blocks.GetCountSinceLatest(5)
	assert.Equal(t, count, int64(3), "There should be 3 blocks!")
}

func TestBlocksGetMissingSinceLatest(t *testing.T) {
	t.Parallel()

	blocks := NewBlocks()
	blocks.AddBlock(&types.Block{Height: 1})
	blocks.AddBlock(&types.Block{Height: 3})
	blocks.AddBlock(&types.Block{Height: 5})

	missing := blocks.GetMissingSinceLatest(5)
	assert.Len(t, missing, 2, "There should be 3 blocks!")
	assert.Contains(t, missing, int64(2), "Blocks mismatch!")
	assert.Contains(t, missing, int64(4), "Blocks mismatch!")
}

func TestBlocksSetBlocks(t *testing.T) {
	t.Parallel()

	blocks := NewBlocks()
	blocks.SetBlocks(map[int64]*types.Block{
		1: {Height: 1},
		2: {Height: 2},
	})

	assert.True(t, blocks.HasBlockAtHeight(1), "Blocks mismatch!")
	assert.True(t, blocks.HasBlockAtHeight(2), "Blocks mismatch!")
}

func TestBlocksTrimBlocksBefore(t *testing.T) {
	t.Parallel()

	blocks := NewBlocks()
	blocks.AddBlock(&types.Block{Height: 1})
	blocks.AddBlock(&types.Block{Height: 2})
	blocks.AddBlock(&types.Block{Height: 3})
	blocks.AddBlock(&types.Block{Height: 4})
	blocks.AddBlock(&types.Block{Height: 5})

	blocks.TrimBefore(3)

	assert.False(t, blocks.HasBlockAtHeight(1), "Blocks mismatch!")
	assert.False(t, blocks.HasBlockAtHeight(2), "Blocks mismatch!")
	assert.False(t, blocks.HasBlockAtHeight(3), "Blocks mismatch!")
	assert.True(t, blocks.HasBlockAtHeight(4), "Blocks mismatch!")
	assert.True(t, blocks.HasBlockAtHeight(5), "Blocks mismatch!")
}
