package responses

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestToBlockInvalid(t *testing.T) {
	t.Parallel()

	blockRaw := &TendermintBlock{
		Header: BlockHeader{Height: "invalid"},
	}

	block, err := blockRaw.ToBlock()
	require.Error(t, err, "Error should be presented!")
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
	require.NoError(t, err, "Error should not be presented!")
	assert.NotNil(t, block, "Block should be presented!")
	assert.Equalf(t, int64(100), block.Height, "Block height mismatch!")
	assert.Len(t, block.Signatures, 2, "Block should have 2 signatures!")
	assert.Equal(t, int32(1), block.Signatures["first"], "Block signature mismatch!")
	assert.Equal(t, int32(2), block.Signatures["second"], "Block signature mismatch!")
}

func TestBlockResponseUnmarshalJson(t *testing.T) {
	t.Parallel()

	successJSON := "{\"jsonrpc\":\"2.0\",\"id\":-1,\"result\":{\"block\":{\"header\":{\"height\":\"12938640\",\"time\":\"2023-09-30T12:31:56.119728652Z\",\"proposer_address\":\"9F478F8D407008B415BA721548A8A2D010254E19\"},\"last_commit\":{\"signatures\":[{\"block_id_flag\":2,\"validator_address\":\"F57E65CB3534A939E1C428241640B9458F6C458D\",\"timestamp\":\"2023-09-30T12:31:56.247075163Z\",\"signature\":\"H/A0W4UnJlDGXpPyOFHu+Yr0nKzECo3HXKdpT6QLt4S7kVptCQiJHdf3dqdVwcEv971HZe7Qt0viiq/toyAlCA==\"}]}}}}"
	errorJSON := "{\"jsonrpc\":\"2.0\",\"id\":-1,\"error\":{\"code\":-32603,\"message\":\"Internal error\",\"data\":\"height 10158584 is not available, lowest height is 12308055\"}}"

	var blockResponse SingleBlockResponse

	err := json.Unmarshal([]byte(successJSON), &blockResponse)

	require.NoError(t, err, "Should not error unmarshalling JSON!")
	assert.Nil(t, blockResponse.Error, "Unmarshall mismatch!")
	assert.NotNil(t, blockResponse.Result, "Unmarshall mismatch!")

	err2 := json.Unmarshal([]byte(errorJSON), &blockResponse)

	require.NoError(t, err2, "Should not error unmarshalling JSON!")
	assert.NotNil(t, blockResponse.Error, "Unmarshall mismatch!")
	assert.Nil(t, blockResponse.Result, "Unmarshall mismatch!")
}
