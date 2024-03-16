package events_test

import (
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
	assert.Equal(
		t,
		"<strong>💡New validator created: <link></strong>",
		rendered,
	)
}

func TestValidatorCreatedFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorCreated{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)
	assert.Equal(
		t,
		"**💡New validator created: <link>**",
		rendered,
	)
}

func TestValidatorCreatedFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorCreated{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeTest, renderData)
	assert.Equal(
		t,
		"Unsupported format type: test",
		rendered,
	)
}
