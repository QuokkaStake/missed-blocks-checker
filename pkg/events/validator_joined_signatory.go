package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorJoinedSignatory struct {
	Validator *types.Validator
}

func (e ValidatorJoinedSignatory) Type() constants.EventName {
	return constants.EventValidatorJoinedSignatory
}

func (e ValidatorJoinedSignatory) GetValidator() *types.Validator {
	return e.Validator
}
