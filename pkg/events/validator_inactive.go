package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorInactive struct {
	Validator *types.Validator
}

func (e ValidatorInactive) Type() string {
	return constants.EventValidatorInactive
}
