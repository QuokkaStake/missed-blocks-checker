package events_test

import (
	"github.com/stretchr/testify/assert"
	"html/template"
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"
)

func TestValidatorActiveBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorActive{Validator: &types.Validator{Moniker: "test"}}

	assert.Equal(t, constants.EventValidatorActive, entry.Type())
	assert.Equal(t, "test", entry.GetValidator().Moniker)
}

func TestValidatorActiveFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorActive{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)

	renderedAsHtml, ok := rendered.(template.HTML)
	assert.True(t, ok)
	assert.Equal(
		t,
		"✅ <strong><link> has joined the active set</strong>notifier1 notifier2",
		string(renderedAsHtml),
	)
}

func TestValidatorActiveFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorActive{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)

	renderedAsString, ok := rendered.(string)
	assert.True(t, ok)
	assert.Equal(
		t,
		"✅ **<link> has joined the active set**notifier1 notifier2",
		renderedAsString,
	)
}

func TestValidatorActiveFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorActive{Validator: &types.Validator{Moniker: "test"}}
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
