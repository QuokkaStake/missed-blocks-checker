package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockHash(t *testing.T) {
	t.Parallel()

	block := Block{Height: 123}
	assert.Equal(t, block.Hash(), "block_123", "Wrong block hash!")
}

func TestBlockSetValidators(t *testing.T) {
	t.Parallel()

	block := Block{Height: 123}
	assert.Nil(t, block.Validators, "Validators should be nil!")
	block.SetValidators(map[string]bool{"1": true})
	assert.NotNil(t, block.Validators, "Validators should not be nil!")
	assert.Len(t, block.Validators, 1, "Validators length should be 1!")
	assert.Equal(t, block.Validators["1"], true, "Validators mismatch!")
}
