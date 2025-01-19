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
func TestReporterJailsFailedToFetch(t *testing.T) {
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

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/jails",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleJailsList(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterJailsOkEmpty(t *testing.T) {
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

	dbClient.Mock.
		ExpectQuery("event, height, validator, payload, time FROM events").
		WillReturnRows(sqlmock.NewRows([]string{"event", "height", "validator", "payload", "time"}))

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/jails",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleJailsList(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterJailsOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytes(assets.GetBytesOrPanic("responses/jails.html")),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{
		Name: "chain",
		TelegramConfig: configPkg.TelegramConfig{
			Token:  "xxx:yyy",
			Chat:   1,
			Admins: []int64{1},
		},
		ExplorerConfig: configPkg.ExplorerConfig{
			ValidatorLinkPattern: "https://example.com/validators/%s",
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
				constants.EventValidatorJailed,
				789,
				"validator1",
				utils.MustJSONMarshall(events.ValidatorJailed{
					Validator: &types.Validator{
						OperatorAddress: "validator1",
						Moniker:         "validator1",
					},
				}),
				renderTime,
			).
			AddRow(
				constants.EventValidatorJailed,
				789,
				"validator2",
				utils.MustJSONMarshall(events.ValidatorJailed{
					Validator: &types.Validator{
						OperatorAddress: "validator2",
						Moniker:         "validator2",
					},
				}),
				renderTime.Add(-time.Hour),
			).
			AddRow(
				constants.EventValidatorTombstoned,
				789,
				"validator3",
				utils.MustJSONMarshall(events.ValidatorTombstoned{
					Validator: &types.Validator{
						OperatorAddress: "validator3",
						Moniker:         "validator3",
					},
				}),
				renderTime.Add(-2*time.Hour),
			),
		)

	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, database)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/jails",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err = reporter.HandleJailsList(ctx)
	require.NoError(t, err)
}
