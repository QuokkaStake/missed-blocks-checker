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

func (e ValidatorInactive) GetValidator() *types.Validator {
	return e.Validator
}
