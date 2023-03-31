package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"main/pkg/constants"
)

type Validator struct {
	Moniker          string
	Description      string
	Identity         string
	SecurityContact  string
	Website          string
	ConsensusAddress string
	OperatorAddress  string
	Commission       float64
	Status           int32
	Jailed           bool
}

func (v *Validator) Active() bool {
	return v.Status == constants.ValidatorBonded
}

func (v *Validator) DetailsChanged(another *Validator) bool {
	return v.Moniker != another.Moniker ||
		v.Description != another.Description ||
		v.Identity != another.Identity ||
		v.SecurityContact != another.SecurityContact ||
		v.Website != another.Website ||
		v.Commission != another.Commission
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

	commission, err := validator.Commission.CommissionRates.Rate.Float64()
	if err != nil {
		panic(err)
	}

	return &Validator{
		Moniker:          validator.Description.Moniker,
		Description:      validator.Description.Details,
		SecurityContact:  validator.Description.SecurityContact,
		Identity:         validator.Description.Identity,
		Website:          validator.Description.Website,
		Commission:       commission,
		ConsensusAddress: fmt.Sprintf("%x", addr),
		OperatorAddress:  validator.OperatorAddress,
		Status:           int32(validator.Status),
		Jailed:           validator.Jailed,
	}
}
