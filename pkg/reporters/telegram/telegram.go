package telegram

import (
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/metrics"
	snapshotPkg "main/pkg/snapshot"
	statePkg "main/pkg/state"
	templatesPkg "main/pkg/templates"
	"main/pkg/types"
	"strings"
	"time"

	"main/pkg/config"
	"main/pkg/utils"

	"github.com/rs/zerolog"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type Reporter struct {
	Token  string
	Chat   int64
	Admins []int64

	Version string

	TelegramBot      *tele.Bot
	Logger           zerolog.Logger
	Config           *config.ChainConfig
	Manager          *statePkg.Manager
	SnapshotManager  *snapshotPkg.Manager
	MetricsManager   *metrics.Manager
	TemplatesManager templatesPkg.Manager

	StopChannel chan bool
}

const (
	MaxMessageSize = 4096
)

func NewReporter(
	chainConfig *config.ChainConfig,
	version string,
	logger zerolog.Logger,
	manager *statePkg.Manager,
	metricsManager *metrics.Manager,
	snapshotManager *snapshotPkg.Manager,
) *Reporter {
	return &Reporter{
		Token:            chainConfig.TelegramConfig.Token,
		Chat:             chainConfig.TelegramConfig.Chat,
		Admins:           chainConfig.TelegramConfig.Admins,
		Config:           chainConfig,
		Logger:           logger.With().Str("component", "telegram_reporter").Logger(),
		Manager:          manager,
		MetricsManager:   metricsManager,
		SnapshotManager:  snapshotManager,
		TemplatesManager: templatesPkg.NewManager(logger, constants.TelegramReporterName),
		Version:          version,
		StopChannel:      make(chan bool),
	}
}

func (reporter *Reporter) Init() {
	if reporter.Token == "" || reporter.Chat == 0 {
		reporter.Logger.Debug().Msg("Telegram credentials not set, not creating Telegram reporter")
		return
	}

	bot, err := tele.NewBot(tele.Settings{
		Token:  reporter.Token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		reporter.Logger.Warn().Err(err).Msg("Could not create Telegram bot")
		return
	}

	if len(reporter.Admins) > 0 {
		reporter.Logger.Debug().Msg("Using admins whitelist")
		bot.Use(middleware.Whitelist(reporter.Admins...))
	}

	queries := []string{
		"help",
		"missing",
		"notifiers",
		"params",
		"status",
		"subscribe",
		"unsubscribe",
		"validators",
	}

	for _, query := range queries {
		reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, query)
	}

	bot.Handle("/start", reporter.HandleHelp)
	bot.Handle("/help", reporter.HandleHelp)
	bot.Handle("/subscribe", reporter.HandleSubscribe)
	bot.Handle("/unsubscribe", reporter.HandleUnsubscribe)
	bot.Handle("/status", reporter.HandleStatus)
	bot.Handle("/validators", reporter.HandleListValidators)
	bot.Handle("/missing", reporter.HandleMissingValidators)
	bot.Handle("/notifiers", reporter.HandleNotifiers)
	bot.Handle("/params", reporter.HandleParams)
	bot.Handle("/config", reporter.HandleParams)

	reporter.TelegramBot = bot
}

func (reporter *Reporter) Start() {
	if reporter.TelegramBot == nil {
		return
	}

	go reporter.TelegramBot.Start()

	<-reporter.StopChannel
	reporter.Logger.Info().Msg("Shutting down...")
	reporter.TelegramBot.Stop()
}

func (reporter *Reporter) Stop() {
	reporter.StopChannel <- true
}

func (reporter *Reporter) Enabled() bool {
	return reporter.Token != "" && reporter.Chat != 0
}

func (reporter *Reporter) GetStateManager() *statePkg.Manager {
	return reporter.Manager
}

func (reporter *Reporter) SerializeEvent(event types.ReportEvent) types.RenderEventItem {
	validator := event.GetValidator()
	notifiers := reporter.Manager.GetNotifiersForReporter(validator.OperatorAddress, constants.TelegramReporterName)

	eventToRender := types.RenderEventItem{
		Event:         event,
		Notifiers:     notifiers,
		ValidatorLink: reporter.Config.ExplorerConfig.GetValidatorLink(validator),
	}

	if eventChanged, ok := event.(events.ValidatorGroupChanged); ok && eventChanged.IsIncreasing() {
		eventToRender.TimeToJail = reporter.Manager.GetTimeTillJail(eventChanged.MissedBlocksAfter)
	}

	return eventToRender
}

func (reporter *Reporter) Send(report *types.Report) error {
	reporter.MetricsManager.LogReport(reporter.Config.Name, report)

	var sb strings.Builder

	for _, event := range report.Events {
		eventToRender := reporter.SerializeEvent(event)
		sb.WriteString(reporter.TemplatesManager.SerializeEvent(eventToRender) + "\n")
	}

	reportString := sb.String()

	reporter.Logger.Trace().Str("report", reportString).Msg("Sending a report")

	if err := reporter.BotSend(reportString); err != nil {
		reporter.Logger.Err(err).Msg("Could not send Telegram message")
		return err
	}
	return nil
}

func (reporter *Reporter) Name() constants.ReporterName {
	return constants.TelegramReporterName
}

func (reporter *Reporter) BotSend(msg string) error {
	messages := utils.SplitStringIntoChunks(msg, MaxMessageSize)

	for _, message := range messages {
		if _, err := reporter.TelegramBot.Send(
			&tele.User{ID: reporter.Chat},
			message,
			tele.ModeHTML,
			tele.NoPreview,
		); err != nil {
			reporter.Logger.Error().Err(err).Msg("Could not send Telegram message")
			return err
		}
	}
	return nil
}

func (reporter *Reporter) BotReply(c tele.Context, msg string) error {
	messages := utils.SplitStringIntoChunks(msg, MaxMessageSize)

	for _, message := range messages {
		if err := c.Reply(strings.TrimSpace(message), tele.ModeHTML, tele.NoPreview); err != nil {
			reporter.Logger.Error().Err(err).Msg("Could not send Telegram message")
			return err
		}
	}
	return nil
}
