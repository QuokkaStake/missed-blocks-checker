package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/google/uuid"
	"main/pkg/logger"
	"time"
)

type Block struct {
	Height     int64
	Time       time.Time
	Proposer   string
	Signatures map[string]int32
}

func (b *Block) Hash() string {
	return fmt.Sprintf("block_%d", b.Height)
}

type WebsocketEmittable interface {
	Hash() string
}

type WSError struct {
	Error error
}

func (w *WSError) Hash() string {
	return "error_" + uuid.NewString()
}

type Validator struct {
	Moniker          string
	ConsensusAddress string
	OperatorAddress  string
	Status           int32
	Jailed           bool
}

func ValidatorFromCosmosValidator(validator stakingTypes.Validator) *Validator {
	interfaceRegistry := types.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	parseCodec := codec.NewProtoCodec(interfaceRegistry)

	if err := validator.UnpackInterfaces(parseCodec); err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not unpack interfaces for validator")
	}

	addr, err := validator.GetConsAddr()
	if err != nil {
		logger.GetDefaultLogger().Fatal().Err(err).Msg("Could not get cons addr for validator")
	}

	return &Validator{
		Moniker:          validator.Description.Moniker,
		ConsensusAddress: fmt.Sprintf("%x", addr),
		OperatorAddress:  validator.OperatorAddress,
		Status:           int32(validator.Status),
		Jailed:           validator.Jailed,
	}
}

type SignatureInto struct {
	Signed      int64
	NoSignature int64
	NotSigned   int64
	Proposed    int64
}
