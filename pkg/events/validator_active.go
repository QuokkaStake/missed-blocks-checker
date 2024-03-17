package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorActive struct {
	Validator *types.Validator
}

func (e ValidatorActive) Type() constants.EventName {
	return constants.EventValidatorActive
}

func (e ValidatorActive) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorActive) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"✅ **%s has joined the active set** %s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"✅ <strong>%s has joined the active set</strong> %s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
