package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorLeftSignatory struct {
	Validator *types.Validator
}

func (e ValidatorLeftSignatory) Type() constants.EventName {
	return constants.EventValidatorLeftSignatory
}

func (e ValidatorLeftSignatory) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorLeftSignatory) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**👋 %s is now not required to sign blocks** %s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>👋 %s is now not required to sign blocks</strong> %s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
