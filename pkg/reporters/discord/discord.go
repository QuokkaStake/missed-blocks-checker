package discord

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/metrics"
	reportPkg "main/pkg/report"
	statePkg "main/pkg/state"
	templatesPkg "main/pkg/templates"
	"main/pkg/types"
	"main/pkg/utils"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type Reporter struct {
	Token   string
	Guild   string
	Channel string

	DiscordSession   *discordgo.Session
	Logger           zerolog.Logger
	Config           *config.ChainConfig
	Manager          *statePkg.Manager
	MetricsManager   *metrics.Manager
	TemplatesManager *templatesPkg.Manager
	Commands         map[string]*Command
}

func NewReporter(
	chainConfig *config.ChainConfig,
	logger zerolog.Logger,
	manager *statePkg.Manager,
	metricsManager *metrics.Manager,
	templatesManager *templatesPkg.Manager,
) *Reporter {
	return &Reporter{
		Token:            chainConfig.DiscordConfig.Token,
		Guild:            chainConfig.DiscordConfig.Guild,
		Channel:          chainConfig.DiscordConfig.Channel,
		Config:           chainConfig,
		Logger:           logger.With().Str("component", "discord_reporter").Logger(),
		Manager:          manager,
		MetricsManager:   metricsManager,
		TemplatesManager: templatesManager,
		Commands:         make(map[string]*Command, 0),
	}
}

func (reporter *Reporter) Init() {
	if !reporter.Enabled() {
		reporter.Logger.Debug().Msg("Discord credentials not set, not creating Discord reporter")
		return
	}
	session, err := discordgo.New("Bot " + reporter.Token)
	if err != nil {
		reporter.Logger.Warn().Err(err).Msg("Error initializing Discord bot")
		return
	}

	reporter.DiscordSession = session

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		reporter.Logger.Warn().Err(err).Msg("Error opening Discord websocket session")
		return
	}

	reporter.Logger.Info().Err(err).Msg("Discord bot listening")

	reporter.Commands = map[string]*Command{
		"params":      reporter.GetParamsCommand(),
		"missing":     reporter.GetMissingCommand(),
		"subscribe":   reporter.GetSubscribeCommand(),
		"unsubscribe": reporter.GetUnubscribeCommand(),
		"status":      reporter.GetStatusCommand(),
		"help":        reporter.GetHelpCommand(),
	}

	for query := range reporter.Commands {
		reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.DiscordReporterName, query)
	}

	go reporter.InitCommands()
}

func (reporter *Reporter) InitCommands() {
	session := reporter.DiscordSession
	var wg sync.WaitGroup
	var mutex sync.Mutex

	session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		commandName := i.ApplicationCommandData().Name

		if command, ok := reporter.Commands[commandName]; ok {
			command.Handler(s, i)
		}
	})

	registeredCommands, err := session.ApplicationCommands(session.State.User.ID, reporter.Guild)
	if err != nil {
		reporter.Logger.Error().Err(err).Msg("Could not fetch registered commands")
		return
	}

	for _, command := range registeredCommands {
		wg.Add(1)
		go func(command *discordgo.ApplicationCommand) {
			defer wg.Done()

			err := session.ApplicationCommandDelete(session.State.User.ID, reporter.Guild, command.ID)
			if err != nil {
				reporter.Logger.Error().Err(err).Str("command", command.Name).Msg("Could not delete command")
				return
			}
			reporter.Logger.Info().Str("command", command.Name).Msg("Deleted command")
		}(command)
	}

	wg.Wait()

	for key, command := range reporter.Commands {
		wg.Add(1)
		go func(key string, command *Command) {
			defer wg.Done()

			cmd, err := session.ApplicationCommandCreate(session.State.User.ID, reporter.Guild, command.Info)
			if err != nil {
				reporter.Logger.Error().Err(err).Str("command", command.Info.Name).Msg("Could not create command")
				return
			}
			reporter.Logger.Info().Str("command", cmd.Name).Msg("Created command")

			mutex.Lock()
			reporter.Commands[key].Info = cmd
			mutex.Unlock()
		}(key, command)
	}

	wg.Wait()
}

func (reporter *Reporter) Enabled() bool {
	return reporter.Token != "" && reporter.Guild != "" && reporter.Channel != ""
}

func (reporter *Reporter) Name() constants.ReporterName {
	return constants.DiscordReporterName
}

func (reporter *Reporter) Send(report *reportPkg.Report) error {
	reporter.MetricsManager.LogReport(reporter.Config.Name, report)

	var sb strings.Builder

	for _, entry := range report.Entries {
		sb.WriteString(reporter.SerializeEntry(entry) + "\n")
	}

	reportString := sb.String()

	reporter.Logger.Trace().Str("report", reportString).Msg("Sending a report")

	_, err := reporter.DiscordSession.ChannelMessageSend(
		reporter.Channel,
		reportString,
	)
	return err
}

func (reporter *Reporter) SerializeEntry(rawEntry reportPkg.Entry) string {
	// validator := rawEntry.GetValidator()
	// notifiers := reporter.Manager.GetNotifiersForReporter(validator.OperatorAddress, reporter.Name())
	// notifiersSerialized := " " + reporter.SerializeNotifiers(notifiers)
	notifiersSerialized := " "

	switch entry := rawEntry.(type) {
	case events.ValidatorGroupChanged:
		timeToJailStr := ""

		if entry.IsIncreasing() {
			timeToJail := reporter.Manager.GetTimeTillJail(entry.MissedBlocksAfter)
			timeToJailStr = fmt.Sprintf(" (%s till jail)", utils.FormatDuration(timeToJail))
		}

		return fmt.Sprintf(
			// a string like "🟡 <validator> is skipping blocks (> 1.0%)  (XXX till jail) <notifier> <notifier2>"
			"**%s %s %s**%s%s",
			entry.GetEmoji(),
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			entry.GetDescription(),
			timeToJailStr,
			notifiersSerialized,
		)
	case events.ValidatorJailed:
		return fmt.Sprintf(
			"**❌ %s was jailed**%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorUnjailed:
		return fmt.Sprintf(
			"**👌 %s was unjailed**%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorInactive:
		return fmt.Sprintf(
			"😔 **%s is now not in the active set**%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorActive:
		return fmt.Sprintf(
			"✅ **%s is now in the active set**%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorTombstoned:
		return fmt.Sprintf(
			"**💀 %s was tombstoned**%s",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
			notifiersSerialized,
		)
	case events.ValidatorCreated:
		return fmt.Sprintf(
			"**💡New validator created: %s**",
			reporter.SerializeLink(reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator)),
		)
	default:
		return fmt.Sprintf("Unsupported event %+v\n", entry)
	}
}

func (reporter *Reporter) SerializeLink(link types.Link) string {
	if link.Href == "" {
		return link.Text
	}

	return fmt.Sprintf("[%s](%s)", link.Text, link.Href)
}

func (reporter *Reporter) BotRespond(s *discordgo.Session, i *discordgo.InteractionCreate, text string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
		},
	})

	if err != nil {
		reporter.Logger.Error().Err(err).Msg("Error sending response")
	}
}
