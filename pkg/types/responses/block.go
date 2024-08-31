package responses

import (
	"encoding/json"
	"main/pkg/types"
	"strconv"
	"time"
)

func (s *SingleBlockResponse) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if result, ok := v["result"]; !ok {
		s.Result = nil
	} else {
		rawBytes, _ := json.Marshal(result) //nolint:errchkjson

		var resultParsed SingleBlockResult
		if err := json.Unmarshal(rawBytes, &resultParsed); err != nil {
			return err
		}

		s.Result = &resultParsed
	}

	if responseError, ok := v["error"]; !ok {
		s.Error = nil
	} else {
		rawBytes, _ := json.Marshal(responseError) //nolint:errchkjson

		var errorParsed ResponseError
		if err := json.Unmarshal(rawBytes, &errorParsed); err != nil {
			return err
		}

		s.Error = &errorParsed
	}

	return nil
}

type SingleBlockResponse struct {
	Result *SingleBlockResult `json:"result"`
	Error  *ResponseError     `json:"error"`
}

type SingleBlockResult struct {
	Block TendermintBlock `json:"block"`
}

type ResponseError struct {
	Data string `json:"data"`
}

type TendermintBlock struct {
	Header     BlockHeader     `json:"header"`
	LastCommit BlockLastCommit `json:"last_commit"`
}

type BlockHeader struct {
	Height   string    `json:"height"`
	Time     time.Time `json:"time"`
	Proposer string    `json:"proposer_address"`
}

type BlockLastCommit struct {
	Signatures []BlockSignature `json:"signatures"`
}

type BlockSignature struct {
	BlockIDFlag      int    `json:"block_id_flag"`
	ValidatorAddress string `json:"validator_address"`
}

func (b *TendermintBlock) ToBlock() (*types.Block, error) {
	height, err := strconv.ParseInt(b.Header.Height, 10, 64)
	if err != nil {
		return nil, err
	}

	signatures := make(map[string]int32, len(b.LastCommit.Signatures))

	for _, signature := range b.LastCommit.Signatures {
		signatures[signature.ValidatorAddress] = int32(signature.BlockIDFlag)
	}

	return &types.Block{
		Height:     height,
		Time:       b.Header.Time,
		Proposer:   b.Header.Proposer,
		Signatures: signatures,
	}, nil
}
