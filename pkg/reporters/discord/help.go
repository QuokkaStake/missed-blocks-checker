package discord

import (
	"main/pkg/constants"

	"github.com/bwmarrin/discordgo"
)

func (reporter *Reporter) GetHelpCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "help",
			Description: "Get the bot help",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.DiscordReporterName, "help")

			template, err := reporter.TemplatesManager.Render("Help", helpRender{
				Version:  reporter.Version,
				Commands: reporter.Commands,
			})
			if err != nil {
				reporter.Logger.Error().Err(err).Str("template", "help").Msg("Error rendering template")
				return
			}

			reporter.BotRespond(s, i, template)
		},
	}
}
