package events_test

import (
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorChangedKeyBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedKey{Validator: &types.Validator{Moniker: "test"}}

	assert.Equal(t, constants.EventValidatorChangedKey, entry.Type())
	assert.Equal(t, "test", entry.GetValidator().Moniker)
}

func TestValidatorChangedKeyFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedKey{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)
	assert.Equal(
		t,
		"<strong>↔️ <link> has changed its signing key</strong> notifier1 notifier2",
		rendered,
	)
}

func TestValidatorChangedKeyFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedKey{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)
	assert.Equal(
		t,
		"**↔️ <link> has changed its signing key** notifier1 notifier2",
		rendered,
	)
}

func TestValidatorChangedKeyFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedKey{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeTest, renderData)
	assert.Equal(
		t,
		"Unsupported format type: test",
		rendered,
	)
}
