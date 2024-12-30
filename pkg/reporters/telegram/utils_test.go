package telegram

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

//nolint:paralleltest // disabled
func TestReplyRenderFailedToRender(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error rendering template: template: pattern matches no files: `telegram/not_found.html`"), //nolint:dupword
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{Name: "chain", TelegramConfig: configPkg.TelegramConfig{Token: "xxx:yyy", Chat: 1}}
	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{})
	stateManager := statePkg.NewManager(*logger, config, metricsManager, nil, nil)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, nil)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/help",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.ReplyRender(ctx, "not_found", nil)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReplyRenderFailedToSend(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		httpmock.NewErrorResponder(errors.New("custom error")))

	config := &configPkg.ChainConfig{Name: "chain", TelegramConfig: configPkg.TelegramConfig{Token: "xxx:yyy", Chat: 1}}
	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{})
	stateManager := statePkg.NewManager(*logger, config, metricsManager, nil, nil)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, nil)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/help",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.ReplyRender(ctx, "not_found", nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}
