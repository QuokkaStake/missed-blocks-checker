package events

import "main/pkg/types"

type ValidatorUnjailed struct {
	Validator *types.Validator
}

func (e ValidatorUnjailed) Type() string {
	return "ValidatorUnjailed"
}
