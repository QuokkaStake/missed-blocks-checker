package events_test

import (
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorChangedMonikerBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedMoniker{Validator: &types.Validator{Moniker: "test"}}

	assert.Equal(t, constants.EventValidatorChangedMoniker, entry.Type())
	assert.Equal(t, "test", entry.GetValidator().Moniker)
}

func TestValidatorChangedMonikerFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedMoniker{
		Validator:    &types.Validator{Moniker: "after"},
		OldValidator: &types.Validator{Moniker: "before"},
	}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)
	assert.Equal(
		t,
		"<strong>✍️ <link> has changed its moniker</strong> (was \"before\") notifier1 notifier2",
		rendered,
	)
}

func TestValidatorChangedMonikerFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedMoniker{
		Validator:    &types.Validator{Moniker: "after"},
		OldValidator: &types.Validator{Moniker: "before"},
	}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)
	assert.Equal(
		t,
		"**✍️ <link> has changed its moniker** (was \"before\") notifier1 notifier2",
		rendered,
	)
}

func TestValidatorChangedMonikerFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedMoniker{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeTest, renderData)
	assert.Equal(
		t,
		"Unsupported format type: test",
		rendered,
	)
}
