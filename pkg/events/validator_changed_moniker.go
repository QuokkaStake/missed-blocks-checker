package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorChangedMoniker struct {
	Validator    *types.Validator
	OldValidator *types.Validator
}

func (e ValidatorChangedMoniker) Type() constants.EventName {
	return constants.EventValidatorChangedMoniker
}

func (e ValidatorChangedMoniker) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorChangedMoniker) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**✍️ %s has changed its moniker** (was \"%s\")%s",
			renderData.ValidatorLink,
			e.OldValidator.Moniker,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>✍️ %s has changed its moniker</strong> (was \"%s\")%s",
			renderData.ValidatorLink,
			e.OldValidator.Moniker,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
