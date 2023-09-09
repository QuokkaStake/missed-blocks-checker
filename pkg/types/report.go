package types

import (
	"main/pkg/constants"
)

type ReportEvent interface {
	Type() constants.EventName
	GetValidator() *Validator
}

type Report struct {
	Events []ReportEvent
}

func (d *Report) Empty() bool {
	return len(d.Events) == 0
}
