package discord

import (
	"main/pkg/constants"

	"github.com/bwmarrin/discordgo"
)

func (reporter *Reporter) GetJailsCountCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "jailscount",
			Description: "See jails count for each validator since the app was started",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.DiscordReporterName, "jailscount")

			snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
			if !found {
				reporter.Logger.Error().Msg("Error searching for historical events!")
				reporter.BotRespond(s, i, "Error searching for jails count!")
				return
			}

			jailsCount, err := reporter.Manager.FindAllJailsCount()
			if err != nil {
				reporter.Logger.Error().Err(err).Msg("Error searching for jails count")
				reporter.BotRespond(s, i, "Error searching for jails count!")
				return
			}

			jailsCountRendered := make([]renderedJailsCount, len(jailsCount))

			for index, validatorJailsCount := range jailsCount {
				validatorEntries := snapshot.Entries.ByValidatorAddresses([]string{validatorJailsCount.Validator})
				if len(validatorEntries) == 0 {
					reporter.BotRespond(s, i, "Validator is not found!")
					return
				}

				jailsCountRendered[index] = renderedJailsCount{
					ValidatorLink: reporter.Config.ExplorerConfig.GetValidatorLink(validatorEntries[0].Validator),
					JailsCount:    validatorJailsCount.JailsCount,
				}
			}

			renderedTemplate, err := reporter.TemplatesManager.Render("JailsCount", jailsCountRendered)
			if err != nil {
				reporter.Logger.Error().Err(err).Msg("Error rendering template")
				return
			}

			reporter.BotRespond(s, i, renderedTemplate)
		},
	}
}
