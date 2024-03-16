package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorJoinedSignatory struct {
	Validator *types.Validator
}

func (e ValidatorJoinedSignatory) Type() constants.EventName {
	return constants.EventValidatorJoinedSignatory
}

func (e ValidatorJoinedSignatory) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorJoinedSignatory) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) any {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**🙋 %s is now required to sign blocks**%s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>🙋 %s is now required to sign blocks</strong>%s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
