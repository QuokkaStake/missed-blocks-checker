package types

import (
	"fmt"
	"main/pkg/constants"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Validator struct {
	Moniker          string
	ConsensusAddress string
	OperatorAddress  string
	Status           int32
	Jailed           bool
}

func (v *Validator) Active() bool {
	return v.Status == constants.ValidatorBonded
}

func ValidatorFromCosmosValidator(validator stakingTypes.Validator) *Validator {
	interfaceRegistry := types.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	parseCodec := codec.NewProtoCodec(interfaceRegistry)

	if err := validator.UnpackInterfaces(parseCodec); err != nil {
		panic(err)
	}

	addr, err := validator.GetConsAddr()
	if err != nil {
		panic(err)
	}

	return &Validator{
		Moniker:          validator.Description.Moniker,
		ConsensusAddress: fmt.Sprintf("%x", addr),
		OperatorAddress:  validator.OperatorAddress,
		Status:           int32(validator.Status),
		Jailed:           validator.Jailed,
	}
}
