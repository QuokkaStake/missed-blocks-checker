package telegram

import (
	"main/assets"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	databasePkg "main/pkg/database"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	"main/pkg/snapshot"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

//nolint:paralleltest // disabled
func TestReporterNotifiersOkEmpty(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Nobody is subscribed to any notifications on chain!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{
		Name: "chain",
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

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/notifiers",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleNotifiers(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterNotifiersOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytes(assets.GetBytesOrPanic("responses/notifiers.html")),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{
		Name: "chain",
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

	stateManager.SetValidators(types.ValidatorsMap{
		"validator1": &types.Validator{OperatorAddress: "validator1", Moniker: "moniker1"},
		"validator2": &types.Validator{OperatorAddress: "validator2", Moniker: "moniker2"},
		"validator3": &types.Validator{OperatorAddress: "validator3", Moniker: "moniker3"},
	})

	stateManager.AddNotifier("validator1", constants.TelegramReporterName, "123", "user1")
	stateManager.AddNotifier("validator1", constants.TelegramReporterName, "456", "user2")
	stateManager.AddNotifier("validator2", constants.TelegramReporterName, "789", "user3")

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/notifiers",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleNotifiers(ctx)
	require.NoError(t, err)
}
