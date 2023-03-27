package events

import "main/pkg/types"

type ValidatorGroupChanged struct {
	Validator          *types.Validator
	MissedBlocksBefore int64
	MissedBlocksAfter  int64
}
