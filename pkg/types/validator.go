package types

import (
	"main/pkg/constants"
)

type SigningInfo struct {
	Tombstoned          bool
	MissedBlocksCounter int64
}

type Validator struct {
	Moniker                 string
	Description             string
	Identity                string
	SecurityContact         string
	Website                 string
	ConsensusAddressHex     string
	ConsensusAddressValcons string
	OperatorAddress         string
	Commission              float64
	Status                  int32
	Jailed                  bool
	SigningInfo             *SigningInfo
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
