package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorUnjailed struct {
	Validator *types.Validator
}

func (e ValidatorUnjailed) Type() constants.EventName {
	return constants.EventValidatorUnjailed
}

func (e ValidatorUnjailed) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorUnjailed) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**👌 %s has been unjailed**%s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>👌 %s has been unjailed</strong>%s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
