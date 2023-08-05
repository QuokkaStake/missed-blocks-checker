package types

import (
	"main/pkg/constants"
	"math/big"
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
	VotingPower             *big.Int
}

func (v *Validator) Active() bool {
	return v.Status == constants.ValidatorBonded
}
