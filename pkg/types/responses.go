package types

import (
	"encoding/json"
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
		rawBytes, err := json.Marshal(result)
		if err != nil {
			return err
		}

		var resultParsed SingleBlockResult
		if err := json.Unmarshal(rawBytes, &resultParsed); err != nil {
			return err
		}

		s.Result = &resultParsed
	}

	if responseError, ok := v["error"]; !ok {
		s.Error = nil
	} else {
		rawBytes, err := json.Marshal(responseError)
		if err != nil {
			return err
		}

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
	Error  *ResponseError     `json:"error,omitempty"`
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

func (b *TendermintBlock) ToBlock() (*Block, error) {
	height, err := strconv.ParseInt(b.Header.Height, 10, 64)
	if err != nil {
		return nil, err
	}

	signatures := make(map[string]int32, len(b.LastCommit.Signatures))

	for _, signature := range b.LastCommit.Signatures {
		signatures[signature.ValidatorAddress] = int32(signature.BlockIDFlag)
	}

	return &Block{
		Height:     height,
		Time:       b.Header.Time,
		Proposer:   b.Header.Proposer,
		Signatures: signatures,
	}, nil
}

type EventResult struct {
	Query string    `json:"query"`
	Data  EventData `json:"data"`
}

type EventData struct {
	Type  string                 `json:"type"`
	Value map[string]interface{} `json:"value"`
}

type AbciQueryResponse struct {
	Result AbciQueryResult `json:"result"`
}

type AbciQueryResult struct {
	Response AbciResponse `json:"response"`
}

type AbciResponse struct {
	Code  int    `json:"code"`
	Log   string `json:"log"`
	Value []byte `json:"value"`
}

type ValidatorsResponse struct {
	Result ValidatorsResult `json:"result"`
	Error  *ResponseError   `json:"error"`
}

type ValidatorsResult struct {
	Validators []HistoricalValidator `json:"validators"`
	Count      string                `json:"count"`
	Total      string                `json:"total"`
}

type HistoricalValidator struct {
	Address string `json:"address"`
}
