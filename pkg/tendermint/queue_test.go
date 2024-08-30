package tendermint

import (
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQueue(t *testing.T) {
	t.Parallel()

	queue := NewQueue(1)

	block := &types.Block{Height: 123}
	require.False(t, queue.Has(block))
	require.Empty(t, queue.Data, 1)

	queue.Add(block)
	require.True(t, queue.Has(block))
	require.Len(t, queue.Data, 1)

	queue.Add(&types.Block{Height: 456})
	require.Len(t, queue.Data, 1)
}
