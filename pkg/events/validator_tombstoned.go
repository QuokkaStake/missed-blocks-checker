package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorTombstoned struct {
	Validator *types.Validator
}

func (e ValidatorTombstoned) Type() constants.EventName {
	return constants.EventValidatorTombstoned
}
