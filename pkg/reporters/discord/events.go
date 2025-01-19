package discord

import (
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/bwmarrin/discordgo"
)

func (reporter *Reporter) GetValidatorEventsCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "events",
			Description: "See latest events for a validator",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "address",
					Description: "Validator address",
					Required:    true,
				},
			},
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.DiscordReporterName, "jails")

			options := i.ApplicationCommandData().Options
			address, _ := options[0].Value.(string)

			snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
			if !found {
				reporter.Logger.Info().Msg("No older snapshot on telegram events query!")
				return
			}

			userEntries := snapshot.Entries.ByValidatorAddresses([]string{address})
			if len(userEntries) == 0 {
				reporter.BotRespond(s, i, "Validator is not found!")
				return
			}

			eventsRaw, err := reporter.Manager.FindLastEventsByValidator(address)
			if err != nil {
				reporter.BotRespond(s, i, "Error searching for historical events!")
				return
			}

			eventsRendered := utils.Map(eventsRaw, func(j types.HistoricalEvent) renderedHistoricalEvent {
				return renderedHistoricalEvent{
					Height: j.Height,
					Time:   j.Time,
					RenderedEvent: reporter.TemplatesManager.SerializeEvent(types.RenderEventItem{
						Event:         j.Event,
						Notifiers:     make(types.Notifiers, 0),
						ValidatorLink: reporter.Config.ExplorerConfig.GetValidatorLink(j.Event.GetValidator()),
					}),
				}
			})

			renderedTemplate, err := reporter.TemplatesManager.Render("Events", eventsRender{
				ValidatorLink: reporter.Config.ExplorerConfig.GetValidatorLink(userEntries[0].Validator),
				Events:        eventsRendered,
			})
			if err != nil {
				reporter.Logger.Error().Err(err).Msg("Error rendering events")
				return
			}

			reporter.BotRespond(s, i, renderedTemplate)
		},
	}
}
