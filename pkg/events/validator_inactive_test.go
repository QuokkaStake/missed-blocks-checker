package events_test

import (
	"html/template"
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorInactiveBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorInactive{Validator: &types.Validator{Moniker: "test"}}

	assert.Equal(t, constants.EventValidatorInactive, entry.Type())
	assert.Equal(t, "test", entry.GetValidator().Moniker)
}

func TestValidatorInactiveFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorInactive{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)

	renderedAsHTML, ok := rendered.(template.HTML)
	assert.True(t, ok)
	assert.Equal(
		t,
		"😔 <strong><link> has left the active set</strong>notifier1 notifier2",
		string(renderedAsHTML),
	)
}

func TestValidatorInactiveFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorInactive{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)

	renderedAsString, ok := rendered.(string)
	assert.True(t, ok)
	assert.Equal(
		t,
		"😔 **<link> has left the active set**notifier1 notifier2",
		renderedAsString,
	)
}

func TestValidatorInactiveFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorInactive{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeTest, renderData)

	renderedAsString, ok := rendered.(string)
	assert.True(t, ok)
	assert.Equal(
		t,
		"Unsupported format type: test",
		renderedAsString,
	)
}
