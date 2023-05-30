package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorCreated struct {
	Validator *types.Validator
}

func (e ValidatorCreated) Type() constants.EventName {
	return constants.EventValidatorCreated
}

func (e ValidatorCreated) GetValidator() *types.Validator {
	return e.Validator
}
