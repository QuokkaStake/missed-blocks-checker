package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorJailed struct {
	Validator *types.Validator
}

func (e ValidatorJailed) Type() constants.EventName {
	return constants.EventValidatorJailed
}

func (e ValidatorJailed) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorJailed) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**❌ %s has been jailed** %s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>❌ %s has been jailed</strong> %s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
