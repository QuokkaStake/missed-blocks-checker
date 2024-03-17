package events_test

import (
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorJailedBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorJailed{Validator: &types.Validator{Moniker: "test"}}

	assert.Equal(t, constants.EventValidatorJailed, entry.Type())
	assert.Equal(t, "test", entry.GetValidator().Moniker)
}

func TestValidatorJailedFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorJailed{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)
	assert.Equal(
		t,
		"<strong>❌ <link> has been jailed</strong> notifier1 notifier2",
		rendered,
	)
}

func TestValidatorJailedFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorJailed{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)
	assert.Equal(
		t,
		"**❌ <link> has been jailed** notifier1 notifier2",
		rendered,
	)
}

func TestValidatorJailedFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorJailed{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeTest, renderData)
	assert.Equal(
		t,
		"Unsupported format type: test",
		rendered,
	)
}
