package telegram

import (
	"errors"
	"main/assets"
	configPkg "main/pkg/config"
	databasePkg "main/pkg/database"
	loggerPkg "main/pkg/logger"
	"main/pkg/metrics"
	"main/pkg/snapshot"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

//nolint:paralleltest // disabled
func TestReporterJailsCountFailedToFetchSnapshot(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error getting validators jails!"),
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
	dbClient := databasePkg.NewStubDatabaseClient()
	database.SetClient(dbClient)

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/jailscount",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleJailsCount(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterJailsCountFailedToFetchEventsCount(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error searching for jails count!"),
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
	dbClient := databasePkg.NewStubDatabaseClient()
	database.SetClient(dbClient)

	snapshotManager.CommitNewSnapshot(100, snapshot.Snapshot{Entries: map[string]*types.Entry{
		"validator": {
			Validator: &types.Validator{OperatorAddress: "validator", Moniker: "validator"},
		},
	}})

	dbClient.Mock.
		ExpectQuery("SELECT validator, count(*) from events").
		WillReturnError(errors.New("custom error"))

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/jailscount",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleJailsCount(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterJailsCountValidatorNotFound(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Validator is not found!"),
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
	dbClient := databasePkg.NewStubDatabaseClient()
	database.SetClient(dbClient)

	snapshotManager.CommitNewSnapshot(100, snapshot.Snapshot{Entries: map[string]*types.Entry{
		"validator1": {
			Validator: &types.Validator{OperatorAddress: "validator", Moniker: "validator"},
		},
	}})

	dbClient.Mock.
		ExpectQuery("SELECT").
		WillReturnRows(sqlmock.
			NewRows([]string{"validator", "count"}).
			AddRow("validator2", 2),
		)

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/jailscount",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleJailsCount(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterJailsCountEmpty(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Nobody has been jailed since the app launch."),
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
	dbClient := databasePkg.NewStubDatabaseClient()
	database.SetClient(dbClient)

	snapshotManager.CommitNewSnapshot(100, snapshot.Snapshot{Entries: map[string]*types.Entry{
		"validator1": {
			Validator: &types.Validator{OperatorAddress: "validator", Moniker: "validator"},
		},
	}})

	dbClient.Mock.
		ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"validator", "count"}))

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/jailscount",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleJailsCount(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterJailsCountOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytes(assets.GetBytesOrPanic("responses/jails-count.html")),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{
		Name: "chain",
		TelegramConfig: configPkg.TelegramConfig{
			Token:  "xxx:yyy",
			Chat:   1,
			Admins: []int64{1},
		},
		ExplorerConfig: configPkg.ExplorerConfig{ValidatorLinkPattern: "https://example.com/validator/%s"},
	}
	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{})
	snapshotManager := snapshot.NewManager(*logger, config, metricsManager)
	database := databasePkg.NewDatabase(*logger, configPkg.DatabaseConfig{})
	dbClient := databasePkg.NewStubDatabaseClient()
	database.SetClient(dbClient)

	snapshotManager.CommitNewSnapshot(100, snapshot.Snapshot{Entries: map[string]*types.Entry{
		"validator1": {
			Validator: &types.Validator{OperatorAddress: "validator1", Moniker: "validator1"},
		},
		"validator2": {
			Validator: &types.Validator{OperatorAddress: "validator2", Moniker: "validator2"},
		},
	}})

	dbClient.Mock.
		ExpectQuery("SELECT").
		WillReturnRows(sqlmock.
			NewRows([]string{"validator", "count"}).
			AddRow("validator2", 2).
			AddRow("validator1", 1),
		)

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/jailscount",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleJailsCount(ctx)
	require.NoError(t, err)
}
