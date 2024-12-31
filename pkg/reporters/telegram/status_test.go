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
func TestReporterStatusNotSubscribed(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("You are not subscribed to any validator's notifications on chain."),
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
			Text:   "/status",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleStatus(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterStatusNoSnapshot(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error getting your validators status!"),
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
		"validator2": &types.Validator{OperatorAddress: "validator1", Moniker: "moniker2"},
		"validator3": &types.Validator{OperatorAddress: "validator1", Moniker: "moniker3"},
		"validator4": &types.Validator{OperatorAddress: "validator1", Moniker: "moniker3"},
	})

	stateManager.AddNotifier("validator1", constants.TelegramReporterName, "123", "user1")
	stateManager.AddNotifier("validator2", constants.TelegramReporterName, "123", "user1")
	stateManager.AddNotifier("validator3", constants.TelegramReporterName, "123", "user1")
	stateManager.AddNotifier("validator4", constants.TelegramReporterName, "123", "user1")

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser", ID: 123},
			Text:   "/status",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleStatus(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterStatusOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytes(assets.GetBytesOrPanic("responses/status.html")),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{
		Name:         "chain",
		BlocksWindow: 100,
		Thresholds:   []float64{0, 5, 100},
		EmojisStart:  []string{"🟢", "🟡"},
		EmojisEnd:    []string{"🟢", "🟡"},
		TelegramConfig: configPkg.TelegramConfig{
			Token:  "xxx:yyy",
			Chat:   1,
			Admins: []int64{1},
		},
	}
	config.RecalculateMissedBlocksGroups()

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
		"validator2": &types.Validator{OperatorAddress: "validator1", Moniker: "moniker2"},
		"validator3": &types.Validator{OperatorAddress: "validator1", Moniker: "moniker3"},
		"validator4": &types.Validator{OperatorAddress: "validator1", Moniker: "moniker3"},
	})

	stateManager.AddNotifier("validator1", constants.TelegramReporterName, "123", "user1")
	stateManager.AddNotifier("validator2", constants.TelegramReporterName, "123", "user1")
	stateManager.AddNotifier("validator3", constants.TelegramReporterName, "123", "user1")
	stateManager.AddNotifier("validator4", constants.TelegramReporterName, "123", "user1")

	snapshotManager.CommitNewSnapshot(123, snapshot.Snapshot{
		Entries: types.Entries{
			"validator1": &types.Entry{
				IsActive:      true,
				Validator:     &types.Validator{OperatorAddress: "validator1", Moniker: "moniker1", VotingPowerPercent: 0.1},
				SignatureInfo: types.SignatureInto{NotSigned: 2},
			},
			"validator2": &types.Entry{
				IsActive:      true,
				Validator:     &types.Validator{OperatorAddress: "validator2", Moniker: "moniker2", Jailed: true},
				SignatureInfo: types.SignatureInto{NotSigned: 25},
			},
			"validator3": &types.Entry{
				IsActive:      false,
				Validator:     &types.Validator{OperatorAddress: "validator3", Moniker: "moniker3"},
				SignatureInfo: types.SignatureInto{NotSigned: 8},
			},
			"validator4": &types.Entry{
				IsActive:      true,
				Validator:     &types.Validator{OperatorAddress: "validator4", Moniker: "moniker4", VotingPowerPercent: 0.5},
				SignatureInfo: types.SignatureInto{NotSigned: 15},
			},
		},
	})

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser", ID: 123},
			Text:   "/status",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleStatus(ctx)
	require.NoError(t, err)
}
