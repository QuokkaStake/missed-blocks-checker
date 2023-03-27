package telegram

import (
	"fmt"
	"html/template"
	"main/pkg/events"
	reportPkg "main/pkg/report"
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
	Templates   map[string]*template.Template
	Config      *config.Config
}

const (
	MaxMessageSize = 4096
)

func NewReporter(
	appConfig *config.Config,
	logger *zerolog.Logger,
) *Reporter {
	return &Reporter{
		Token:     appConfig.TelegramConfig.Token,
		Chat:      appConfig.TelegramConfig.Chat,
		Admins:    appConfig.TelegramConfig.Admins,
		Config:    appConfig,
		Logger:    logger.With().Str("component", "telegram_reporter").Logger(),
		Templates: make(map[string]*template.Template, 0),
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
	//bot.Handle("/aliases", reporter.HandleGetAliases)

	reporter.TelegramBot = bot
	go reporter.TelegramBot.Start()
}

func (reporter *Reporter) Enabled() bool {
	return reporter.Token != "" && reporter.Chat != 0
}

func (reporter *Reporter) SerializeEntry(rawEntry reportPkg.ReportEntry) template.HTML {
	switch entry := rawEntry.(type) {
	case events.ValidatorGroupChanged:
		return template.HTML(fmt.Sprintf(
			"%s is skipping blocks (%d -> %d)",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			entry.MissedBlocksBefore,
			entry.MissedBlocksAfter,
		))
	case events.ValidatorJailed:
		return template.HTML(fmt.Sprintf(
			"%s is jailed",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
		))
	case events.ValidatorUnjailed:
		return template.HTML(fmt.Sprintf(
			"%s is unjailed",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
		))
	default:
		return template.HTML(fmt.Sprintf("Unsupported event %+v\n", entry))
	}
}

func (reporter *Reporter) Send(report *reportPkg.Report) error {
	var sb strings.Builder

	for _, entry := range report.Entries {
		sb.WriteString(string(reporter.SerializeEntry(entry) + "\n"))
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

func (reporter *Reporter) SerializeDate(date time.Time) template.HTML {
	return template.HTML(date.Format(time.RFC822))
}

func (reporter *Reporter) SerializeLink(link types.Link) template.HTML {
	if link.Href == "" {
		return template.HTML(link.Text)
	}

	return template.HTML(fmt.Sprintf("<a href='%s'>%s</a>", link.Href, link.Text))
}
