package types

import (
	"main/pkg/constants"
)

type ReportEntry interface {
	Type() constants.EventName
	GetValidator() *Validator
}

type Report struct {
	Entries []ReportEntry
}

func (d *Report) Empty() bool {
	return len(d.Entries) == 0
}
