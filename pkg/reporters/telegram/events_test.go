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
	"main/pkg/utils"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	tele "gopkg.in/telebot.v3"
)

//nolint:paralleltest // disabled
func TestReporterEventsInvalidInvocation(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Usage: /events &lt;validator address&gt;"),
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
			Text:   "/events",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleValidatorEventsList(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterEventsErrorFetchingSnapshot(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error getting validator events!"),
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
			Text:   "/events validator",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleValidatorEventsList(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterEventsNotSubscribed(t *testing.T) {
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

	snapshotManager.CommitNewSnapshot(100, snapshot.Snapshot{Entries: map[string]*types.Entry{}})

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/events validator",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleValidatorEventsList(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterEventsErrorFetchingEvents(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error searching for historical events!"),
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

	dbClient.Mock.
		ExpectQuery("event, height, validator, payload, time FROM events").
		WillReturnError(errors.New("custom error"))

	snapshotManager.CommitNewSnapshot(100, snapshot.Snapshot{Entries: map[string]*types.Entry{
		"validator": {
			Validator: &types.Validator{OperatorAddress: "validator", Moniker: "validator"},
		},
	}})

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/events validator",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleValidatorEventsList(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterEventsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytes(assets.GetBytesOrPanic("responses/events.html")),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{
		Name:           "chain",
		ExplorerConfig: configPkg.ExplorerConfig{ValidatorLinkPattern: "https://example.com/validators/%s"},
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

	renderTime, err := time.Parse(time.RFC3339, "2025-01-19T11:03:00Z")
	require.NoError(t, err)

	dbClient.Mock.
		ExpectQuery("event, height, validator, payload, time FROM events").
		WillReturnRows(sqlmock.
			NewRows([]string{"event", "height", "validator", "payload", "time"}).
			AddRow(
				constants.EventValidatorUnjailed,
				789,
				"validator",
				utils.MustJSONMarshall(events.ValidatorUnjailed{
					Validator: &types.Validator{
						OperatorAddress: "validator",
						Moniker:         "validator",
					},
				}),
				renderTime,
			).
			AddRow(
				constants.EventValidatorJailed,
				789,
				"validator",
				utils.MustJSONMarshall(events.ValidatorJailed{
					Validator: &types.Validator{
						OperatorAddress: "validator",
						Moniker:         "validator",
					},
				}),
				renderTime.Add(-time.Hour),
			),
		)

	snapshotManager.CommitNewSnapshot(100, snapshot.Snapshot{Entries: map[string]*types.Entry{
		"validator": {
			Validator: &types.Validator{OperatorAddress: "validator", Moniker: "validator"},
		},
	}})

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/events validator",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err = reporter.HandleValidatorEventsList(ctx)
	require.NoError(t, err)
}
