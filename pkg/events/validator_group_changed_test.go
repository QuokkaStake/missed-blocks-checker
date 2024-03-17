package events_test

import (
	configPkg "main/pkg/config"
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorGroupChangedBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorGroupChanged{Validator: &types.Validator{Moniker: "test"}}

	assert.Equal(t, constants.EventValidatorGroupChanged, entry.Type())
	assert.Equal(t, "test", entry.GetValidator().Moniker)
}

func TestValidatorGetDescriptionAndEmojiIncreasing(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorGroupChanged{
		Validator: &types.Validator{Moniker: "test"},
		MissedBlocksGroupBefore: &configPkg.MissedBlocksGroup{
			Start:      0,
			End:        5,
			DescStart:  "start1",
			DescEnd:    "end1",
			EmojiStart: "emojistart1",
			EmojiEnd:   "emojiend1",
		},
		MissedBlocksGroupAfter: &configPkg.MissedBlocksGroup{
			Start:      6,
			End:        10,
			DescStart:  "start2",
			DescEnd:    "end2",
			EmojiStart: "emojistart2",
			EmojiEnd:   "emojiend2",
		},
	}

	assert.Equal(t, "start2", entry.GetDescription())
	assert.Equal(t, "emojistart2", entry.GetEmoji())
}

func TestValidatorGetDescriptionAndEmojiDecreasing(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorGroupChanged{
		Validator: &types.Validator{Moniker: "test"},
		MissedBlocksGroupAfter: &configPkg.MissedBlocksGroup{
			Start:      0,
			End:        5,
			DescStart:  "start1",
			DescEnd:    "end1",
			EmojiStart: "emojistart1",
			EmojiEnd:   "emojiend1",
		},
		MissedBlocksGroupBefore: &configPkg.MissedBlocksGroup{
			Start:      6,
			End:        10,
			DescStart:  "start2",
			DescEnd:    "end2",
			EmojiStart: "emojistart2",
			EmojiEnd:   "emojiend2",
		},
	}

	assert.Equal(t, "end1", entry.GetDescription())
	assert.Equal(t, "emojiend1", entry.GetEmoji())
}

func TestValidatorGroupChangedFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorGroupChanged{
		Validator: &types.Validator{Moniker: "test"},
		MissedBlocksGroupAfter: &configPkg.MissedBlocksGroup{
			Start:      0,
			End:        5,
			DescStart:  "start1",
			DescEnd:    "end1",
			EmojiStart: "emojistart1",
			EmojiEnd:   "emojiend1",
		},
		MissedBlocksGroupBefore: &configPkg.MissedBlocksGroup{
			Start:      6,
			End:        10,
			DescStart:  "start2",
			DescEnd:    "end2",
			EmojiStart: "emojistart2",
			EmojiEnd:   "emojiend2",
		},
	}

	renderData := types.ReportEventRenderData{
		Notifiers:     "notifier1 notifier2",
		ValidatorLink: "<link>",
		TimeToJail:    "<jail>",
	}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)
	assert.Equal(
		t,
		"<strong>emojiend1 <link> end1</strong><jail> notifier1 notifier2",
		rendered,
	)
}

func TestValidatorGroupChangedFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorGroupChanged{
		Validator: &types.Validator{Moniker: "test"},
		MissedBlocksGroupAfter: &configPkg.MissedBlocksGroup{
			Start:      0,
			End:        5,
			DescStart:  "start1",
			DescEnd:    "end1",
			EmojiStart: "emojistart1",
			EmojiEnd:   "emojiend1",
		},
		MissedBlocksGroupBefore: &configPkg.MissedBlocksGroup{
			Start:      6,
			End:        10,
			DescStart:  "start2",
			DescEnd:    "end2",
			EmojiStart: "emojistart2",
			EmojiEnd:   "emojiend2",
		},
	}

	renderData := types.ReportEventRenderData{
		Notifiers:     "notifier1 notifier2",
		ValidatorLink: "<link>",
		TimeToJail:    "<jail>",
	}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)
	assert.Equal(
		t,
		"**emojiend1 <link> end1**<jail> notifier1 notifier2",
		rendered,
	)
}

func TestValidatorGroupChangedFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorGroupChanged{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeTest, renderData)
	assert.Equal(
		t,
		"Unsupported format type: test",
		rendered,
	)
}
