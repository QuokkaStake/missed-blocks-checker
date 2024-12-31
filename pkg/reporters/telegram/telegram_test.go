package telegram

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	databasePkg "main/pkg/database"
	"main/pkg/events"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	"main/pkg/snapshot"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"strings"
	"testing"
	"time"

	tele "gopkg.in/telebot.v3"

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
	require.Equal(t, constants.TelegramReporterName, reporter.Name())
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

//nolint:paralleltest // disabled
func TestReporterBotReplyFail(t *testing.T) {
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

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/help",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.BotReply(ctx, strings.Repeat("a", 5000))
	require.Error(t, err)
}

//nolint:paralleltest // disabled
func TestReporterSendFail(t *testing.T) {
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

	config := &configPkg.ChainConfig{
		Name:               "chain",
		BlocksWindow:       1000,
		MinSignedPerWindow: 10,
		TelegramConfig: configPkg.TelegramConfig{
			Token:  "xxx:yyy",
			Chat:   1,
			Admins: []int64{1},
		},
	}
	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{})
	stateManager := statePkg.NewManager(*logger, config, metricsManager, nil, nil)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, nil)
	reporter.Init()

	err := reporter.Send(&types.Report{
		Events: []types.ReportEvent{
			events.ValidatorGroupChanged{
				MissedBlocksBefore:      10,
				MissedBlocksAfter:       100,
				MissedBlocksGroupBefore: &configPkg.MissedBlocksGroup{},
				MissedBlocksGroupAfter:  &configPkg.MissedBlocksGroup{},
				Validator:               &types.Validator{OperatorAddress: "validator", Moniker: "moniker"},
			},
		},
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest // disabled
func TestReporterSendOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("<strong> moniker is skipping blocks</strong> (1 hour 10 minutes 50 seconds till jail)"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{
		Name:               "chain",
		BlocksWindow:       1000,
		MinSignedPerWindow: 0.05,
		TelegramConfig: configPkg.TelegramConfig{
			Token:  "xxx:yyy",
			Chat:   1,
			Admins: []int64{1},
		},
	}

	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{})
	snapshotManager := snapshot.NewManager(*logger, config, metricsManager)
	database := databasePkg.NewDatabase(*logger, configPkg.DatabaseConfig{})
	database.SetClient(&databasePkg.StubDatabaseClient{})

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	currentTime := time.Now()

	err := stateManager.AddBlock(&types.Block{Height: 1, Time: currentTime})
	require.NoError(t, err)

	err = stateManager.AddBlock(&types.Block{Height: 2, Time: currentTime.Add(5 * time.Second)})
	require.NoError(t, err)

	err = reporter.Send(&types.Report{
		Events: []types.ReportEvent{
			events.ValidatorGroupChanged{
				MissedBlocksBefore:      10,
				MissedBlocksAfter:       100,
				MissedBlocksGroupBefore: &configPkg.MissedBlocksGroup{Start: 10, DescStart: "is skipping blocks"},
				MissedBlocksGroupAfter:  &configPkg.MissedBlocksGroup{Start: 100, DescStart: "is skipping blocks"},
				Validator:               &types.Validator{OperatorAddress: "validator", Moniker: "moniker"},
			},
		},
	})
	require.NoError(t, err)
}
