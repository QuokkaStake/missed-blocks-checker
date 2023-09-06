package discord

import (
	"main/pkg/constants"

	"github.com/bwmarrin/discordgo"
)

func (reporter *Reporter) GetParamsCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "params",
			Description: "Get the bot params and chain info",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.DiscordReporterName, "params")

			blockTime := reporter.Manager.GetBlockTime()
			maxTimeToJail := reporter.Manager.GetTimeTillJail(0)

			template, err := reporter.TemplatesManager.Render("Params", paramsRender{
				Config:        reporter.Config,
				BlockTime:     blockTime,
				MaxTimeToJail: maxTimeToJail,
			})
			if err != nil {
				reporter.Logger.Error().Err(err).Msg("Error rendering params template")
				return
			}

			if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: template,
				},
			}); err != nil {
				reporter.Logger.Error().Err(err).Msg("Error sending params")
			}
		},
	}
}
