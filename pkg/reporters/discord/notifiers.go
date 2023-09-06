package discord

import (
	"main/pkg/constants"

	"github.com/bwmarrin/discordgo"
)

func (reporter *Reporter) GetNotifiersCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "notifiers",
			Description: "Get notifiers for all validators on chain",
			Version:     "0",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.DiscordReporterName, "notifiers")

			validators := reporter.Manager.GetValidators().ToSlice()
			entries := make([]notifierEntry, 0)

			for _, validator := range validators {
				link := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
				notifiers := reporter.Manager.GetNotifiersForReporter(validator.OperatorAddress, constants.DiscordReporterName)
				if len(notifiers) == 0 {
					continue
				}

				entries = append(entries, notifierEntry{
					Link:      link,
					Notifiers: notifiers,
				})
			}

			template, err := reporter.TemplatesManager.Render("Notifiers", notifierRender{
				Entries: entries,
				Config:  reporter.Config,
			})
			if err != nil {
				reporter.BotRespond(s, i, "Error rendering notifiers template")
				return
			}

			reporter.BotRespond(s, i, template)
		},
	}
}
