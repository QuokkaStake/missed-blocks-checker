package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorChangedCommission struct {
	Validator    *types.Validator
	OldValidator *types.Validator
}

func (e ValidatorChangedCommission) Type() constants.EventName {
	return constants.EventValidatorChangedCommission
}

func (e ValidatorChangedCommission) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorChangedCommission) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**💰️ %s has changed its commission**: %.2f%% -> %.2f%% %s",
			renderData.ValidatorLink,
			e.OldValidator.Commission*100,
			e.Validator.Commission*100,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>💰️ %s has changed its commission</strong>: %.2f%% -> %.2f%% %s",
			renderData.ValidatorLink,
			e.OldValidator.Commission*100,
			e.Validator.Commission*100,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
