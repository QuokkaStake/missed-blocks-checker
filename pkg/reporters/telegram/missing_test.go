package telegram

import (
	"main/assets"
	configPkg "main/pkg/config"
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
func TestReporterMissingFailedToFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasText("Error getting validators list!"),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{Name: "chain", TelegramConfig: configPkg.TelegramConfig{
		Token:  "xxx:yyy",
		Chat:   1,
		Admins: []int64{1},
	}}
	logger := loggerPkg.GetNopLogger()
	metricsManager := metrics.NewManager(*logger, configPkg.MetricsConfig{})
	snapshotManager := snapshot.NewManager(*logger, config, metricsManager)
	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, nil)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/missing",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleMissingValidators(ctx)
	require.NoError(t, err)
}

//nolint:paralleltest // disabled
func TestReporterMissingOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/getMe",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-bot-ok.json")))

	httpmock.RegisterMatcherResponder(
		"POST",
		"https://api.telegram.org/botxxx:yyy/sendMessage",
		types.TelegramResponseHasBytes(assets.GetBytesOrPanic("responses/missing.html")),
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("telegram-send-message-ok.json")),
	)

	config := &configPkg.ChainConfig{
		Name:         "chain",
		BlocksWindow: 100,
		Thresholds:   []float64{0, 10, 100},
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
	stateManager := statePkg.NewManager(*logger, config, metricsManager, snapshotManager, nil)
	reporter := NewReporter(config, "1.2.3", *logger, stateManager, metricsManager, snapshotManager)
	reporter.Init()

	snapshotManager.CommitNewSnapshot(123, snapshot.Snapshot{
		Entries: types.Entries{
			"validator1": &types.Entry{
				IsActive:      true,
				Validator:     &types.Validator{Moniker: "moniker1"},
				SignatureInfo: types.SignatureInto{NotSigned: 5},
			},
			"validator2": &types.Entry{
				IsActive:      true,
				Validator:     &types.Validator{Moniker: "moniker2"},
				SignatureInfo: types.SignatureInto{NotSigned: 25},
			},
			"validator3": &types.Entry{
				IsActive:      false,
				Validator:     &types.Validator{Moniker: "moniker3"},
				SignatureInfo: types.SignatureInto{NotSigned: 8},
			},
			"validator4": &types.Entry{
				IsActive:      true,
				Validator:     &types.Validator{Moniker: "moniker4"},
				SignatureInfo: types.SignatureInto{NotSigned: 15},
			},
		},
	})

	ctx := reporter.TelegramBot.NewContext(tele.Update{
		ID: 1,
		Message: &tele.Message{
			Sender: &tele.User{Username: "testuser"},
			Text:   "/missing",
			Chat:   &tele.Chat{ID: 2},
		},
	})

	err := reporter.HandleMissingValidators(ctx)
	require.NoError(t, err)
}
