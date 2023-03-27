package events

import "main/pkg/types"

type ValidatorJailed struct {
	Validator *types.Validator
}
