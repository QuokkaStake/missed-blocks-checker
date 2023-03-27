package telegram

import (
	"fmt"
	"html"
	"main/pkg/events"
	reportPkg "main/pkg/report"
	statePkg "main/pkg/state"
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

	TelegramBot *tele.Bot
	Logger      zerolog.Logger
	Config      *config.Config
	Manager     *statePkg.Manager
}

const (
	MaxMessageSize = 4096
)

func NewReporter(
	appConfig *config.Config,
	logger *zerolog.Logger,
	manager *statePkg.Manager,
) *Reporter {
	return &Reporter{
		Token:   appConfig.TelegramConfig.Token,
		Chat:    appConfig.TelegramConfig.Chat,
		Admins:  appConfig.TelegramConfig.Admins,
		Config:  appConfig,
		Logger:  logger.With().Str("component", "telegram_reporter").Logger(),
		Manager: manager,
	}
}

func (reporter *Reporter) Init() {
	if reporter.Token == "" || reporter.Chat == 0 {
		reporter.Logger.Debug().Msg("Telegram credentials not set, not creating Telegram reporter.")
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

	//bot.Handle("/help", reporter.HandleHelp)
	//bot.Handle("/start", reporter.HandleHelp)
	//bot.Handle("/status", reporter.HandleListNodesStatus)
	//bot.Handle("/config", reporter.HandleGetConfig)
	//bot.Handle("/alias", reporter.HandleSetAlias)
	bot.Handle("/subscribe", reporter.HandleSubscribe)

	reporter.TelegramBot = bot
	go reporter.TelegramBot.Start()
}

func (reporter *Reporter) Enabled() bool {
	return reporter.Token != "" && reporter.Chat != 0
}

func (reporter *Reporter) SerializeEntry(rawEntry reportPkg.ReportEntry) string {
	switch entry := rawEntry.(type) {
	case events.ValidatorGroupChanged:
		return fmt.Sprintf(
			"<strong>%s %s</strong> %s",
			entry.GetEmoji(),
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			html.EscapeString(entry.GetDescription()),
		)
	case events.ValidatorJailed:
		return fmt.Sprintf(
			"%s is jailed",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
		)
	case events.ValidatorUnjailed:
		return fmt.Sprintf(
			"%s is unjailed",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
		)
	default:
		return fmt.Sprintf("Unsupported event %+v\n", entry)
	}
}

func (reporter *Reporter) Send(report *reportPkg.Report) error {
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

func (reporter *Reporter) Name() string {
	return "telegram-reporter"
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
		if err := c.Reply(message, tele.ModeHTML); err != nil {
			reporter.Logger.Error().Err(err).Msg("Could not send Telegram message")
			return err
		}
	}
	return nil
}

func (reporter *Reporter) SerializeDate(date time.Time) string {
	return date.Format(time.RFC822)
}

func (reporter *Reporter) SerializeLink(link types.Link) string {
	if link.Href == "" {
		return link.Text
	}

	return fmt.Sprintf("<a href='%s'>%s</a>", link.Href, link.Text)
}
