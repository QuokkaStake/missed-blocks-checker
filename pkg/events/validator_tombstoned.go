package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorTombstoned struct {
	Validator *types.Validator
}

func (e ValidatorTombstoned) Type() constants.EventName {
	return constants.EventValidatorTombstoned
}

func (e ValidatorTombstoned) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorTombstoned) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**💀 %s has been tombstoned**%s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>💀 %s has been tombstoned</strong>%s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
