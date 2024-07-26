package telegram

import (
	"errors"
	"fmt"
	"html"
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

	reporter.AddCommand("/start", bot, reporter.GetHelpCommand())
	reporter.AddCommand("/help", bot, reporter.GetHelpCommand())
	reporter.AddCommand("/subscribe", bot, reporter.GetSubscribeCommand())
	reporter.AddCommand("/unsubscribe", bot, reporter.GetUnsubscribeCommand())
	reporter.AddCommand("/status", bot, reporter.GetStatusCommand())
	reporter.AddCommand("/validators", bot, reporter.GetListValidatorsCommand())
	reporter.AddCommand("/missing", bot, reporter.GetMissingValidatorsCommand())
	reporter.AddCommand("/notifiers", bot, reporter.GetNotifiersCommand())
	reporter.AddCommand("/params", bot, reporter.GetParamsCommand())
	reporter.AddCommand("/config", bot, reporter.GetParamsCommand())

	reporter.TelegramBot = bot
	go reporter.TelegramBot.Start()
}

func (reporter *Reporter) AddCommand(query string, bot *tele.Bot, command Command) {
	bot.Handle(query, func(c tele.Context) error {
		reporter.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Str("command", command.Name).
			Msg("Got query")

		reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, command.Name)

		args := strings.Split(c.Text(), " ")

		if len(args)-1 < command.MinArgs {
			if err := reporter.BotReply(c, html.EscapeString(fmt.Sprintf(command.Usage, args[0]))); err != nil {
				return err
			}

			return errors.New("invalid invocation")
		}

		result, err := command.Execute(c)
		if err != nil {
			reporter.Logger.Error().
				Err(err).
				Str("command", command.Name).
				Msg("Error processing command")
			if result != "" {
				return reporter.BotReply(c, result)
			} else {
				return reporter.BotReply(c, "Internal error!")
			}
		}

		return reporter.BotReply(c, result)
	})
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
			&tele.User{
				ID: reporter.Chat,
			},
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
		if err := c.Reply(message, tele.ModeHTML, tele.NoPreview); err != nil {
			reporter.Logger.Error().Err(err).Msg("Could not send Telegram message")
			return err
		}
	}
	return nil
}

func (reporter *Reporter) SerializeDate(date time.Time) string {
	return date.Format(time.RFC822)
}
