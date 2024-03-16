package events

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorChangedKey struct {
	Validator    *types.Validator
	OldValidator *types.Validator
}

func (e ValidatorChangedKey) Type() constants.EventName {
	return constants.EventValidatorChangedKey
}

func (e ValidatorChangedKey) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorChangedKey) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**↔️ %s has changed its signing key**%s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>↔️ %s has changed its signing key</strong>%s",
			renderData.ValidatorLink,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
