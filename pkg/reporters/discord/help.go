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

			template, err := reporter.TemplatesManager.Render("Help", reporter.Commands)
			if err != nil {
				reporter.Logger.Error().Err(err).Str("template", "help").Msg("Error rendering template")
				return
			}

			if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: template,
				},
			}); err != nil {
				reporter.Logger.Error().Err(err).Msg("Error sending help")
			}
		},
	}
}
