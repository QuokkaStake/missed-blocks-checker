package types

import (
	"cosmossdk.io/math"
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
	Jailed                  bool
	SigningInfo             *SigningInfo

	VotingPower                  math.LegacyDec
	VotingPowerPercent           float64
	CumulativeVotingPowerPercent float64
	Rank                         int
}
