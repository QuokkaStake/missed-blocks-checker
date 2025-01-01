package templates

import (
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewManagerTelegram(t *testing.T) {
	t.Parallel()

	manager := NewManager(*loggerPkg.GetNopLogger(), constants.TelegramReporterName)
	require.IsType(t, &TelegramTemplateManager{}, manager)
}

func TestNewManagerDiscord(t *testing.T) {
	t.Parallel()

	manager := NewManager(*loggerPkg.GetNopLogger(), constants.DiscordReporterName)
	require.IsType(t, &DiscordTemplateManager{}, manager)
}

func TestNewManagerNotValid(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	NewManager(*loggerPkg.GetNopLogger(), constants.TestReporterName)
}
