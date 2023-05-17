package events

import "main/pkg/types"

type ValidatorJailed struct {
	Validator *types.Validator
}

func (e ValidatorJailed) Type() string {
	return "ValidatorJailed"
}
