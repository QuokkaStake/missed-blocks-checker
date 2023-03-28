package events

import "main/pkg/types"

type ValidatorInactive struct {
	Validator *types.Validator
}
