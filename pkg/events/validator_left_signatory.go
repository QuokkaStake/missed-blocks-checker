package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorLeftSignatory struct {
	Validator *types.Validator
}

func (e ValidatorLeftSignatory) Type() constants.EventName {
	return constants.EventValidatorLeftSignatory
}

func (e ValidatorLeftSignatory) GetValidator() *types.Validator {
	return e.Validator
}
