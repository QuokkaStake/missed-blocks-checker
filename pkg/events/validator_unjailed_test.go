package events_test

import (
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorUnjailedBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorUnjailed{Validator: &types.Validator{Moniker: "test"}}

	assert.Equal(t, constants.EventValidatorUnjailed, entry.Type())
	assert.Equal(t, "test", entry.GetValidator().Moniker)
}

func TestValidatorUnjailedFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorUnjailed{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)
	assert.Equal(
		t,
		"<strong>👌 <link> has been unjailed</strong>notifier1 notifier2",
		rendered,
	)
}

func TestValidatorUnjailedFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorUnjailed{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)
	assert.Equal(
		t,
		"**👌 <link> has been unjailed**notifier1 notifier2",
		rendered,
	)
}

func TestValidatorUnjailedFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorUnjailed{Validator: &types.Validator{Moniker: "test"}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeTest, renderData)
	assert.Equal(
		t,
		"Unsupported format type: test",
		rendered,
	)
}
