package types

import (
	"fmt"
	"main/pkg/utils"
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
	Height string    `json:"height"`
	Time   time.Time `json:"time"`
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
		panic(fmt.Sprintf("Could not convert block height to string: %s", b.Header.Height))
	}

	return &Block{
		Height: height,
		Time:   b.Header.Time,
		Signatures: utils.Map(b.LastCommit.Signatures, func(s BlockSignature) Signature {
			return s.ToSignature(height)
		}),
	}
}

func (s *BlockSignature) ToSignature(height int64) Signature {
	return Signature{
		ValidatorAddr: s.ValidatorAddress,
		Height:        height,
		Signed:        s.BlockIDFlag == 2,
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
