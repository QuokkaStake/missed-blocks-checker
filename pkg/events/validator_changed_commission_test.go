package events_test

import (
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatorChangedCommissionBase(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedCommission{Validator: &types.Validator{Commission: 0.01}}

	assert.Equal(t, constants.EventValidatorChangedCommission, entry.Type())
	assert.Equal(t, 0.01, entry.GetValidator().Commission)
}

func TestValidatorChangedCommissionFormatHTML(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedCommission{
		Validator:    &types.Validator{Commission: 0.02},
		OldValidator: &types.Validator{Commission: 0.01},
	}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeHTML, renderData)
	assert.Equal(
		t,
		"<strong>💰️ <link> has changed its commission</strong>: 1.00% -> 2.00% notifier1 notifier2",
		rendered,
	)
}

func TestValidatorChangedCommissionFormatMarkdown(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedCommission{
		Validator:    &types.Validator{Commission: 0.02},
		OldValidator: &types.Validator{Commission: 0.01},
	}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeMarkdown, renderData)
	assert.Equal(
		t,
		"**💰️ <link> has changed its commission**: 1.00% -> 2.00% notifier1 notifier2",
		rendered,
	)
}

func TestValidatorChangedCommissionFormatUnsupported(t *testing.T) {
	t.Parallel()

	entry := events.ValidatorChangedCommission{Validator: &types.Validator{Commission: 0.01}}
	renderData := types.ReportEventRenderData{Notifiers: "notifier1 notifier2", ValidatorLink: "<link>"}
	rendered := entry.Render(constants.FormatTypeTest, renderData)
	assert.Equal(
		t,
		"Unsupported format type: test",
		rendered,
	)
}
