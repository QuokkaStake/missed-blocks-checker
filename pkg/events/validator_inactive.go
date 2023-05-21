package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorInactive struct {
	Validator *types.Validator
}

func (e ValidatorInactive) Type() constants.EventName {
	return constants.EventValidatorInactive
}
