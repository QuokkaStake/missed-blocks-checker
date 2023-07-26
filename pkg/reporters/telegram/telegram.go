package telegram

import (
	"fmt"
	"html"
	"html/template"
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/metrics"
	reportPkg "main/pkg/report"
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

	TelegramBot      *tele.Bot
	Logger           zerolog.Logger
	Config           *config.ChainConfig
	Manager          *statePkg.Manager
	MetricsManager   *metrics.Manager
	TemplatesManager *templatesPkg.Manager
}

const (
	MaxMessageSize = 4096
)

func NewReporter(
	chainConfig *config.ChainConfig,
	logger zerolog.Logger,
	manager *statePkg.Manager,
	metricsManager *metrics.Manager,
	templatesManager *templatesPkg.Manager,
) *Reporter {
	return &Reporter{
		Token:            chainConfig.TelegramConfig.Token,
		Chat:             chainConfig.TelegramConfig.Chat,
		Admins:           chainConfig.TelegramConfig.Admins,
		Config:           chainConfig,
		Logger:           logger.With().Str("component", "telegram_reporter").Logger(),
		Manager:          manager,
		MetricsManager:   metricsManager,
		TemplatesManager: templatesManager,
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
	go reporter.TelegramBot.Start()
}

func (reporter *Reporter) Enabled() bool {
	return reporter.Token != "" && reporter.Chat != 0
}

func (reporter *Reporter) SerializeEntry(rawEntry reportPkg.Entry) string {
	validator := rawEntry.GetValidator()
	notifiers := reporter.Manager.GetNotifiersForReporter(validator.OperatorAddress, reporter.Name())
	notifiersSerialized := " " + reporter.SerializeNotifiers(notifiers)

	switch entry := rawEntry.(type) {
	case events.ValidatorGroupChanged:
		timeToJailStr := ""

		if entry.IsIncreasing() {
			timeToJail := reporter.Manager.GetTimeTillJail(entry.MissedBlocksAfter)
			timeToJailStr = fmt.Sprintf(" (%s till jail)", utils.FormatDuration(timeToJail))
		}

		return fmt.Sprintf(
			// a string like "🟡 <validator> is skipping blocks (> 1.0%)  (XXX till jail) <notifier> <notifier2>"
			"<strong>%s %s %s</strong>%s%s",
			entry.GetEmoji(),
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			html.EscapeString(entry.GetDescription()),
			timeToJailStr,
			notifiersSerialized,
		)
	case events.ValidatorJailed:
		return fmt.Sprintf(
			"<strong>❌ %s was jailed</strong>%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorUnjailed:
		return fmt.Sprintf(
			"<strong>👌 %s was unjailed</strong>%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorInactive:
		return fmt.Sprintf(
			"😔 <strong>%s is now not in the active set</strong>%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorActive:
		return fmt.Sprintf(
			"✅ <strong>%s is now in the active set</strong>%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorTombstoned:
		return fmt.Sprintf(
			"<strong>💀 %s was tombstoned</strong>%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorCreated:
		return fmt.Sprintf(
			"<strong>💡New validator created: %s</strong>",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
		)
	default:
		return fmt.Sprintf("Unsupported event %+v\n", entry)
	}
}

func (reporter *Reporter) Send(report *reportPkg.Report) error {
	reporter.MetricsManager.LogReport(reporter.Config.Name, report)

	var sb strings.Builder

	for _, entry := range report.Entries {
		sb.WriteString(reporter.SerializeEntry(entry) + "\n")
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

func (reporter *Reporter) SerializeLink(link types.Link) template.HTML {
	if link.Href == "" {
		return template.HTML(link.Text)
	}

	return template.HTML(fmt.Sprintf("<a href='%s'>%s</a>", link.Href, link.Text))
}

func (reporter *Reporter) SerializeNotifiers(notifiers []*types.Notifier) string {
	notifiersNormalized := utils.Map(notifiers, reporter.SerializeNotifier)

	return strings.Join(notifiersNormalized, " ")
}

func (reporter *Reporter) SerializeNotifier(notifier *types.Notifier) string {
	if strings.HasPrefix(notifier.UserName, "@") {
		return notifier.UserName
	}

	return "@" + notifier.UserName
}
