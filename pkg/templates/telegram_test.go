package templates

import (
	"html/template"
	loggerPkg "main/pkg/logger"
	"main/pkg/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelegramTemplateRenderNotFound(t *testing.T) {
	t.Parallel()

	manager := NewTelegramTemplateManager(*loggerPkg.GetNopLogger())
	result, err := manager.Render("not-found", nil)

	require.Error(t, err)
	require.Empty(t, result)
}

func TestTelegramTemplateRenderFailedToRender(t *testing.T) {
	t.Parallel()

	manager := NewTelegramTemplateManager(*loggerPkg.GetNopLogger())
	result, err := manager.Render("Status", "test")

	require.Error(t, err)
	require.Empty(t, result)
}

func TestTelegramTemplateRenderOk(t *testing.T) {
	t.Parallel()

	manager := NewTelegramTemplateManager(*loggerPkg.GetNopLogger())
	result, err := manager.Render("Help", "1.2.3")
	require.NoError(t, err)
	require.NotEmpty(t, result)

	result2, err2 := manager.Render("Help", "1.2.3")
	require.NoError(t, err2)
	require.NotEmpty(t, result2)
}

func TestTelegramGetHTMLTemplateWrongType(t *testing.T) {
	t.Parallel()

	manager := NewTelegramTemplateManager(*loggerPkg.GetNopLogger())
	manager.Templates["templateName"] = "test"

	_, err := manager.GetHTMLTemplate("templateName")
	require.Error(t, err)
	require.ErrorContains(t, err, "error converting template")
}

func TestTelegramSerialize(t *testing.T) {
	t.Parallel()

	manager := NewTelegramTemplateManager(*loggerPkg.GetNopLogger())

	testTime, err := time.Parse(time.RFC3339, "2024-12-31T15:49:45Z")
	require.NoError(t, err)

	require.Equal(t, "31 Dec 24 15:49 UTC", manager.SerializeDate(testTime))
	require.Equal(t, template.HTML("text"), manager.SerializeLink(types.Link{
		Text: "text",
	}))
	require.Equal(t, template.HTML("<a href='https://example.com'>text</a>"), manager.SerializeLink(types.Link{
		Text: "text",
		Href: "https://example.com",
	}))
	require.Equal(t, "@user1 @user2", manager.SerializeNotifiers(types.Notifiers{
		{UserName: "user1"},
		{UserName: "@user2"},
	}))
}
