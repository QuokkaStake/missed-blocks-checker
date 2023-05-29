package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToBlockInvalid(t *testing.T) {
	t.Parallel()

	blockRaw := &TendermintBlock{
		Header: BlockHeader{Height: "invalid"},
	}

	block, err := blockRaw.ToBlock()
	assert.NotNil(t, err, "Error should be presented!")
	assert.Nil(t, block, "Block should not be presented!")
}

func TestToBlockValid(t *testing.T) {
	t.Parallel()

	blockRaw := &TendermintBlock{
		Header: BlockHeader{Height: "100"},
		LastCommit: BlockLastCommit{
			Signatures: []BlockSignature{
				{ValidatorAddress: "first", BlockIDFlag: 1},
				{ValidatorAddress: "second", BlockIDFlag: 2},
			},
		},
	}

	block, err := blockRaw.ToBlock()
	assert.Nil(t, err, "Error should not be presented!")
	assert.NotNil(t, block, "Block should be presented!")
	assert.Equalf(t, block.Height, int64(100), "Block height mismatch!")
	assert.Len(t, block.Signatures, 2, "Block should have 2 signatures!")
	assert.Equal(t, block.Signatures["first"], int32(1), "Block signature mismatch!")
	assert.Equal(t, block.Signatures["second"], int32(2), "Block signature mismatch!")
}
