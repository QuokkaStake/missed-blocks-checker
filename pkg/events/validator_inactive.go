package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorInactive struct {
	Validator *types.Validator
}

func (e ValidatorInactive) Type() constants.EventName {
	return constants.EventValidatorInactive
}

func (e ValidatorInactive) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorInactive) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"😔 **%s has left the active set** %s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"😔 <strong>%s has left the active set</strong> %s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
