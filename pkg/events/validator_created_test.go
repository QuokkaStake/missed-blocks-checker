package events_test

import (
	"html/template"
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorCreatedBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorCreated{Validator: &types.Validator{Moniker: "test"}}

	assert.Equal(t, constants.EventValidatorCreated, entry.Type())
	assert.Equal(t, "test", entry.GetValidator().Moniker)
}

func TestValidatorCreatedFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorCreated{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)

	renderedAsHTML, ok := rendered.(template.HTML)
	assert.True(t, ok)
	assert.Equal(
		t,
		"<strong>💡New validator created: <link></strong>",
		string(renderedAsHTML),
	)
}

func TestValidatorCreatedFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorCreated{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)

	renderedAsString, ok := rendered.(string)
	assert.True(t, ok)
	assert.Equal(
		t,
		"**💡New validator created: <link>**",
		renderedAsString,
	)
}

func TestValidatorCreatedFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorCreated{Validator: &types.Validator{Moniker: "test"}}
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
