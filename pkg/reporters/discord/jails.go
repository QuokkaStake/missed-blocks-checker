package discord

import (
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/utils"

	"github.com/bwmarrin/discordgo"
)

func (reporter *Reporter) GetJailsCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "jails",
			Description: "See latest jails and tombstones events",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.DiscordReporterName, "jails")

			jailsRaw, err := reporter.Manager.FindLastEventsByType([]constants.EventName{
				constants.EventValidatorJailed,
				constants.EventValidatorTombstoned,
			})
			if err != nil {
				reporter.Logger.Error().Err(err).Msg("Error searching for historical events!")
				return
			}

			jailsRendered := utils.Map(jailsRaw, func(j types.HistoricalEvent) renderedHistoricalEvent {
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

			renderedTemplate, err := reporter.TemplatesManager.Render("Jails", jailsRendered)
			if err != nil {
				reporter.Logger.Error().Err(err).Msg("Error rendering missing")
				return
			}

			reporter.BotRespond(s, i, renderedTemplate)
		},
	}
}
