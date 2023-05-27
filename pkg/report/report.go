package report

import (
	"main/pkg/constants"
	"main/pkg/types"
)

type Entry interface {
	Type() constants.EventName
	GetValidator() *types.Validator
}

type Report struct {
	Entries []Entry
}

func (d *Report) Empty() bool {
	return len(d.Entries) == 0
}
