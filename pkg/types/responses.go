package types

import (
	"strconv"
	"time"
)

type SingleBlockResponse struct {
	Result SingleBlockResult `json:"result"`
}

type SingleBlockResult struct {
	Block TendermintBlock `json:"block"`
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

type BlockSearchResponse struct {
	Result BlockSearchResult `json:"result"`
}

type BlockSearchResult struct {
	Blocks []SingleBlockResult `json:"blocks"`
}

func (b *TendermintBlock) ToBlock() *Block {
	height, err := strconv.ParseInt(b.Header.Height, 10, 64)
	if err != nil {
		panic(err)
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
	}
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
	Value []byte `json:"value"`
}

type ValidatorsResponse struct {
	Result ValidatorsResult `json:"result"`
}

type ValidatorsResult struct {
	Validators []ActiveSetValidator `json:"validators"`
}

type ActiveSetValidator struct {
	Address string `json:"address"`
}
