package converter

import (
	"fmt"
	"main/pkg/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	slashingTypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Converter struct {
	registry   codecTypes.InterfaceRegistry
	parseCodec *codec.ProtoCodec
}

func NewConverter() *Converter {
	interfaceRegistry := codecTypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	parseCodec := codec.NewProtoCodec(interfaceRegistry)

	return &Converter{
		registry:   interfaceRegistry,
		parseCodec: parseCodec,
	}
}

func (c *Converter) GetConsensusAddress(validator stakingTypes.Validator) string {
	if err := validator.UnpackInterfaces(c.parseCodec); err != nil {
		panic(err)
	}

	addr, err := validator.GetConsAddr()
	if err != nil {
		panic(err)
	}

	return addr.String()
}

func (c *Converter) ValidatorFromCosmosValidator(
	validator stakingTypes.Validator,
	signingInfo *slashingTypes.ValidatorSigningInfo,
) *types.Validator {
	if err := validator.UnpackInterfaces(c.parseCodec); err != nil {
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

	var valSigningInfo *types.SigningInfo

	if signingInfo != nil {
		valSigningInfo = &types.SigningInfo{
			MissedBlocksCounter: signingInfo.MissedBlocksCounter,
			Tombstoned:          signingInfo.Tombstoned,
		}
	}

	return &types.Validator{
		Moniker:                 validator.Description.Moniker,
		Description:             validator.Description.Details,
		SecurityContact:         validator.Description.SecurityContact,
		Identity:                validator.Description.Identity,
		Website:                 validator.Description.Website,
		Commission:              commission,
		ConsensusAddressHex:     fmt.Sprintf("%x", addr),
		ConsensusAddressValcons: addr.String(),
		OperatorAddress:         validator.OperatorAddress,
		Jailed:                  validator.Jailed,
		SigningInfo:             valSigningInfo,
		VotingPower:             validator.DelegatorShares,
	}
}

func (c *Converter) SetValidatorConsumerConsensusAddr(validator *types.Validator, consumerKey string) error {
	consAddress, err := sdkTypes.ConsAddressFromBech32(consumerKey)
	if err != nil {
		return err
	}

	validator.ConsensusAddressValcons = consAddress.String()
	validator.ConsensusAddressHex = fmt.Sprintf("%x", consAddress)

	return nil
}

func (c *Converter) GetConsKeyFromBech32(addr string) (string, error) {
	consAddress, err := sdkTypes.ConsAddressFromBech32(addr)
	if err != nil {
		return "", err
	}

	return consAddress.String(), nil
}
