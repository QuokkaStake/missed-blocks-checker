package events

import (
	"fmt"
	"html"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/types"
)

type ValidatorGroupChanged struct {
	Validator               *types.Validator
	MissedBlocksBefore      int64
	MissedBlocksAfter       int64
	MissedBlocksGroupBefore *configPkg.MissedBlocksGroup
	MissedBlocksGroupAfter  *configPkg.MissedBlocksGroup
}

func (e ValidatorGroupChanged) Type() constants.EventName {
	return constants.EventValidatorGroupChanged
}

func (e ValidatorGroupChanged) GetDescription() string {
	// increasing
	if e.IsIncreasing() {
		return e.MissedBlocksGroupAfter.DescStart
	}

	// decreasing
	return e.MissedBlocksGroupAfter.DescEnd
}

func (e ValidatorGroupChanged) GetEmoji() string {
	// increasing
	if e.IsIncreasing() {
		return e.MissedBlocksGroupAfter.EmojiStart
	}

	// decreasing
	return e.MissedBlocksGroupAfter.EmojiEnd
}

func (e ValidatorGroupChanged) IsIncreasing() bool {
	return e.MissedBlocksGroupBefore.Start < e.MissedBlocksGroupAfter.Start
}

func (e ValidatorGroupChanged) GetValidator() *types.Validator {
	return e.Validator
}

func (e ValidatorGroupChanged) Render(formatType constants.FormatType, renderData types.ReportEventRenderData) string {
	// a string like "🟡 <validator> (link) is skipping blocks (> 1.0%)  (XXX till jail) <notifier> <notifier2>"
	switch formatType {
	case constants.FormatTypeMarkdown:
		return fmt.Sprintf(
			"**%s %s %s**%s%s",
			e.GetEmoji(),
			renderData.ValidatorLink,
			e.GetDescription(),
			renderData.TimeToJail,
			renderData.Notifiers,
		)
	case constants.FormatTypeHTML:
		return fmt.Sprintf(
			"<strong>%s %s %s</strong>%s%s",
			e.GetEmoji(),
			renderData.ValidatorLink,
			html.EscapeString(e.GetDescription()),
			renderData.TimeToJail,
			renderData.Notifiers,
		)
	default:
		return fmt.Sprintf("Unsupported format type: %s", formatType)
	}
}
