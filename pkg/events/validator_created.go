package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorCreated struct {
	Validator *types.Validator
}

func (e ValidatorCreated) Type() constants.EventName {
	return constants.EventValidatorCreated
}

func (e ValidatorCreated) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorCreated) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**💡New validator created: %s**",
			renderData.ValidatorLink,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>💡New validator created: %s</strong>",
			renderData.ValidatorLink,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
