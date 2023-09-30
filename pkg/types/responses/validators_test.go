package responses

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorsResponseUnmarshalJson(t *testing.T) {
	t.Parallel()

	successJSON := "{\"jsonrpc\":\"2.0\",\"id\":-1,\"result\":{\"block_height\":\"12939374\",\"validators\":[{\"address\":\"F57E65CB3534A939E1C428241640B9458F6C458D\",\"pub_key\":{\"type\":\"tendermint/PubKeyEd25519\",\"value\":\"yeayC1a/lvThcKJFMW+icYvE+MzPh/pIVMDZDgF/3cc=\"},\"voting_power\":\"1206754697\",\"proposer_priority\":\"8824167576\"}],\"count\":\"30\",\"total\":\"80\"}}"
	errorJSON := "{\"jsonrpc\":\"2.0\",\"id\":-1,\"error\":{\"code\":-32603,\"message\":\"Internal error\",\"data\":\"height 1 is not available, lowest height is 12308055\"}}"

	var validatorsResponse ValidatorsResponse

	err := json.Unmarshal([]byte(successJSON), &validatorsResponse)

	assert.Nil(t, err, "Should not error unmarshalling JSON!")
	assert.Nil(t, validatorsResponse.Error, "Unmarshall mismatch!")
	assert.NotNil(t, validatorsResponse.Result, "Unmarshall mismatch!")

	err2 := json.Unmarshal([]byte(errorJSON), &validatorsResponse)

	assert.Nil(t, err2, "Should not error unmarshalling JSON!")
	assert.NotNil(t, validatorsResponse.Error, "Unmarshall mismatch!")
	assert.Nil(t, validatorsResponse.Result, "Unmarshall mismatch!")
}
