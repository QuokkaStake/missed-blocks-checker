package types

import (
	"main/pkg/constants"
)

type ReportEventRenderData struct {
	Notifiers     string
	ValidatorLink string
	TimeToJail    string
}

type ReportEvent interface {
	Type() constants.EventName
	GetValidator() *Validator
	Render(formatType constants.FormatType, renderData ReportEventRenderData) any
}

type Report struct {
	Events []ReportEvent
}

func (d *Report) Empty() bool {
	return len(d.Events) == 0
}
