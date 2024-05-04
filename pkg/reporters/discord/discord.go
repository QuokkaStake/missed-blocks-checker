package discord

import (
	"main/pkg/config"
	"main/pkg/constants"
	"main/pkg/events"
	"main/pkg/metrics"
	snapshotPkg "main/pkg/snapshot"
	statePkg "main/pkg/state"
	templatesPkg "main/pkg/templates"
	types "main/pkg/types"
	"main/pkg/utils"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

type Reporter struct {
	Token   string
	Guild   string
	Channel string

	Version string

	DiscordSession   *discordgo.Session
	Logger           zerolog.Logger
	Config           *config.ChainConfig
	Manager          *statePkg.Manager
	MetricsManager   *metrics.Manager
	SnapshotManager  *snapshotPkg.Manager
	TemplatesManager templatesPkg.Manager
	Commands         map[string]*Command
}

func NewReporter(
	chainConfig *config.ChainConfig,
	version string,
	logger zerolog.Logger,
	manager *statePkg.Manager,
	metricsManager *metrics.Manager,
	snapshotManager *snapshotPkg.Manager,
) *Reporter {
	return &Reporter{
		Token:            chainConfig.DiscordConfig.Token,
		Guild:            chainConfig.DiscordConfig.Guild,
		Channel:          chainConfig.DiscordConfig.Channel,
		Config:           chainConfig,
		Logger:           logger.With().Str("component", "discord_reporter").Logger(),
		Manager:          manager,
		MetricsManager:   metricsManager,
		SnapshotManager:  snapshotManager,
		TemplatesManager: templatesPkg.NewManager(logger, constants.DiscordReporterName),
		Commands:         make(map[string]*Command, 0),
		Version:          version,
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
		"unsubscribe": reporter.GetUnsubscribeCommand(),
		"status":      reporter.GetStatusCommand(),
		"help":        reporter.GetHelpCommand(),
		"notifiers":   reporter.GetNotifiersCommand(),
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

func (reporter *Reporter) SerializeEvent(event types.ReportEvent) types.RenderEventItem {
	validator := event.GetValidator()
	notifiers := reporter.Manager.GetNotifiersForReporter(validator.OperatorAddress, constants.DiscordReporterName)

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

	_, err := reporter.DiscordSession.ChannelMessageSend(
		reporter.Channel,
		reportString,
	)
	return err
}

func (reporter *Reporter) BotRespond(s *discordgo.Session, i *discordgo.InteractionCreate, text string) {
	chunks := utils.SplitStringIntoChunks(text, 2000)
	firstChunk, rest := chunks[0], chunks[1:]

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: firstChunk,
		},
	}); err != nil {
		reporter.Logger.Error().Err(err).Msg("Error sending response")
	}

	for index, chunk := range rest {
		if _, err := s.FollowupMessageCreate(i.Interaction, false, &discordgo.WebhookParams{
			Content: chunk,
		}); err != nil {
			reporter.Logger.Error().
				Int("chunk", index).
				Err(err).
				Msg("Error sending followup message")
		}
	}
}

func (reporter *Reporter) SerializeDate(date time.Time) string {
	return date.Format(time.RFC822)
}
