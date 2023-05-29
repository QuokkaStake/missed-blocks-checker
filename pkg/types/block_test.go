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
