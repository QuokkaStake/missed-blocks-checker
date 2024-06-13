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

			snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
			if !found {
				reporter.Logger.Info().Msg("No older snapshot on telegram params query!")
				reporter.BotRespond(s, i, "Error getting params")
				return
			}

			activeValidators := snapshot.Entries.GetActive()
			var amount int
			if reporter.Config.IsConsumer.Bool {
				_, amount = snapshot.Entries.GetSoftOutOutThreshold(reporter.Config.ConsumerSoftOptOut)
			}

			template, err := reporter.TemplatesManager.Render("Params", paramsRender{
				Config:                   reporter.Config,
				BlockTime:                blockTime,
				MaxTimeToJail:            maxTimeToJail,
				ConsumerOptOutValidators: amount,
				ValidatorsCount:          len(activeValidators),
			})
			if err != nil {
				reporter.Logger.Error().Err(err).Msg("Error rendering params template")
				return
			}

			reporter.BotRespond(s, i, template)
		},
	}
}
