package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorJailed struct {
	Validator *types.Validator
}

func (e ValidatorJailed) Type() constants.EventName {
	return constants.EventValidatorJailed
}

func (e ValidatorJailed) GetValidator() *types.Validator {
	return e.Validator
}
