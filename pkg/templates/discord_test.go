package templates

import (
	"html/template"
	configPkg "main/pkg/config"
	"main/pkg/events"
	loggerPkg "main/pkg/logger"
	"main/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDiscordTemplateRenderNotFound(t *testing.T) {
	t.Parallel()

	manager := NewDiscordTemplateManager(*loggerPkg.GetNopLogger())
	result, err := manager.Render("not-found", nil)

	require.Error(t, err)
	require.Empty(t, result)
}

func TestDiscordTemplateRenderFailedToRender(t *testing.T) {
	t.Parallel()

	manager := NewDiscordTemplateManager(*loggerPkg.GetNopLogger())
	result, err := manager.Render("Status", "test")

	require.Error(t, err)
	require.Empty(t, result)
}

func TestDiscordTemplateRenderOk(t *testing.T) {
	t.Parallel()

	testStruct := struct {
		Version  string
		Commands map[string]interface{}
	}{
		Version:  "1.2.3",
		Commands: map[string]interface{}{},
	}

	manager := NewDiscordTemplateManager(*loggerPkg.GetNopLogger())
	result, err := manager.Render("Help", testStruct)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	result2, err2 := manager.Render("Help", testStruct)
	require.NoError(t, err2)
	require.NotEmpty(t, result2)
}

func TestDiscordGetTemplateWrongType(t *testing.T) {
	t.Parallel()

	manager := NewDiscordTemplateManager(*loggerPkg.GetNopLogger())
	manager.Templates["templateName"] = "test"

	_, err := manager.GetTemplate("templateName")
	require.Error(t, err)
	require.ErrorContains(t, err, "error converting template")
}

func TestDiscordSerialize(t *testing.T) {
	t.Parallel()

	manager := NewDiscordTemplateManager(*loggerPkg.GetNopLogger())

	testTime, err := time.Parse(time.RFC3339, "2024-12-31T15:49:45Z")
	require.NoError(t, err)

	require.Equal(t, "31 Dec 24 15:49 UTC", manager.SerializeDate(testTime))
	require.Equal(t, template.HTML("text"), manager.SerializeLink(types.Link{
		Text: "text",
	}))
	require.Equal(t, template.HTML("[text](<https://example.com>)"), manager.SerializeLink(types.Link{
		Text: "text",
		Href: "https://example.com",
	}))
	require.Equal(t, "<@user1> <@user2>", manager.SerializeNotifiers(types.Notifiers{
		{UserID: "user1"},
		{UserID: "user2"},
	}))
	require.Equal(t, "`@user1` `@user2`", manager.SerializeNotifiersNoLinks(types.Notifiers{
		{UserName: "user1"},
		{UserName: "user2"},
	}))
}

func TestDiscordSerializeEvent(t *testing.T) {
	t.Parallel()

	manager := NewDiscordTemplateManager(*loggerPkg.GetNopLogger())
	manager.Templates["templateName"] = "test"

	result := manager.SerializeEvent(types.RenderEventItem{
		ValidatorLink: types.Link{Text: "moniker"},
		TimeToJail:    10 * time.Second,
		Event: events.ValidatorGroupChanged{
			MissedBlocksBefore:      10,
			MissedBlocksAfter:       100,
			MissedBlocksGroupBefore: &configPkg.MissedBlocksGroup{Start: 10, DescStart: "is skipping blocks"},
			MissedBlocksGroupAfter:  &configPkg.MissedBlocksGroup{Start: 100, DescStart: "is skipping blocks"},
			Validator:               &types.Validator{OperatorAddress: "validator", Moniker: "moniker"},
		},
	})
	require.Equal(t, "** moniker is skipping blocks** (10 seconds till jail) ", result)
}

func TestDiscordSerializeEventAnother(t *testing.T) {
	t.Parallel()

	manager := NewDiscordTemplateManager(*loggerPkg.GetNopLogger())
	manager.Templates["templateName"] = "test"

	result := manager.SerializeEvent(types.RenderEventItem{
		ValidatorLink: types.Link{Text: "moniker", Href: "https://example.com"},
		TimeToJail:    10 * time.Second,
		Event: events.ValidatorGroupChanged{
			MissedBlocksBefore:      10,
			MissedBlocksAfter:       100,
			MissedBlocksGroupBefore: &configPkg.MissedBlocksGroup{Start: 10, DescStart: "is skipping blocks"},
			MissedBlocksGroupAfter:  &configPkg.MissedBlocksGroup{Start: 100, DescStart: "is skipping blocks"},
			Validator:               &types.Validator{OperatorAddress: "validator", Moniker: "moniker"},
		},
	})
	require.Equal(t, "** moniker ([link](<https://example.com>)) is skipping blocks** (10 seconds till jail) ", result)
}
