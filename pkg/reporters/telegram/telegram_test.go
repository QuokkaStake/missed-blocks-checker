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
)

//nolint:paralleltest // disabled
func TestReporterInitNoCredentials(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	config := &configPkg.ChainConfig{Name: "chain", TelegramConfig: configPkg.TelegramConfig{}}
	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{})
	stateManager := statePkg.NewManager(*logger, config, metricsManager, nil, nil)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, nil)
	reporter.Init()
	reporter.Start()

	require.False(t, reporter.Enabled())
	require.NotEmpty(t, reporter.GetStateManager())
}

//nolint:paralleltest // disabled
func TestReporterInitFailedToFetchBot(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewErrorResponder(errors.New("custom error")))

	config := &configPkg.ChainConfig{Name: "chain", TelegramConfig: configPkg.TelegramConfig{Token: "xxx:yyy", Chat: 1}}
	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{})
	stateManager := statePkg.NewManager(*logger, config, metricsManager, nil, nil)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, nil)
	reporter.Init()
}

//nolint:paralleltest // disabled
func TestReporterInitOk(t *testing.T) {
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

	config := &configPkg.ChainConfig{Name: "chain", TelegramConfig: configPkg.TelegramConfig{
		Token:  "xxx:yyy",
		Chat:   1,
		Admins: []int64{1},
	}}
	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{})
	stateManager := statePkg.NewManager(*logger, config, metricsManager, nil, nil)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, nil)
	reporter.Init()

	go reporter.Start()
	reporter.Stop()
}
