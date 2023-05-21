package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorUnjailed struct {
	Validator *types.Validator
}

func (e ValidatorUnjailed) Type() constants.EventName {
	return constants.EventValidatorUnjailed
}
