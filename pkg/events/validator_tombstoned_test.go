package events_test

import (
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorTombstonedBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorTombstoned{Validator: &types.Validator{Moniker: "test"}}

	assert.Equal(t, constants.EventValidatorTombstoned, entry.Type())
	assert.Equal(t, "test", entry.GetValidator().Moniker)
}

func TestValidatorTombstonedFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorTombstoned{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)
	assert.Equal(
		t,
		"<strong>💀 <link> has been tombstoned</strong>notifier1 notifier2",
		rendered,
	)
}

func TestValidatorTombstonedFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorTombstoned{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)
	assert.Equal(
		t,
		"**💀 <link> has been tombstoned**notifier1 notifier2",
		rendered,
	)
}

func TestValidatorTombstonedFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorTombstoned{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeTest, renderData)
	assert.Equal(
		t,
		"Unsupported format type: test",
		rendered,
	)
}
