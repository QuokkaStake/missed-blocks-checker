package events

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorActive struct {
	Validator *types.Validator
}

func (e ValidatorActive) Type() string {
	return constants.EventValidatorActive
}
