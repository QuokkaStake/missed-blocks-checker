package events

import "main/pkg/types"

type ValidatorInactive struct {
	Validator *types.Validator
}

func (e ValidatorInactive) Type() string {
	return "ValidatorInactive"
}
