package types

import (
	htmlTemplate "html/template"
	"main/pkg/constants"
)

type ReportEventRenderData struct {
	Notifiers     string
	ValidatorLink htmlTemplate.HTML
	TimeToJail    string
}

type ReportEvent interface {
	Type() constants.EventName
	GetValidator() *Validator
	Render(formatType constants.FormatType, renderData ReportEventRenderData) string
}

type Report struct {
	Events []ReportEvent
}

func (d *Report) Empty() bool {
	return len(d.Events) == 0
}
