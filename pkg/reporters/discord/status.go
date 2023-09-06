package discord

import (
	"fmt"
	"main/pkg/constants"

	"github.com/bwmarrin/discordgo"
)

func (reporter *Reporter) GetStatusCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "status",
			Description: "See the status of the validators you are subscribed to",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.DiscordReporterName, "status")

			user := i.User
			if user == nil {
				user = i.Member.User
			}
			if user == nil {
				reporter.BotRespond(s, i, "Could not fetch user!")
				return
			}

			operatorAddresses := reporter.Manager.GetValidatorsForNotifier(reporter.Name(), user.ID)
			if len(operatorAddresses) == 0 {
				reporter.BotRespond(s, i, fmt.Sprintf(
					"You are not subscribed to any validator's notifications on %s.",
					reporter.Config.GetName(),
				))
				return
			}

			entries := make([]statusEntry, len(operatorAddresses))

			for index, operatorAddress := range operatorAddresses {
				validator, found := reporter.Manager.GetValidator(operatorAddress)
				if !found {
					reporter.BotRespond(s, i, fmt.Sprintf(
						"Could not find a validator with address <code>%s</code> on %s",
						operatorAddress,
						reporter.Config.GetName(),
					))
					return
				}

				entries[index] = statusEntry{
					Validator: validator,
					Link:      reporter.Config.ExplorerConfig.GetValidatorLink(validator),
				}

				if validator.Active() && !validator.Jailed {
					signatureInfo, err := reporter.Manager.GetValidatorMissedBlocks(validator)
					entries[index].Error = err
					entries[index].SigningInfo = signatureInfo
				}
			}

			template, err := reporter.TemplatesManager.Render("Status", statusRender{
				ChainConfig: reporter.Config,
				Entries:     entries,
			})
			if err != nil {
				reporter.BotRespond(s, i, "Could not render template")
				return
			}
			reporter.BotRespond(s, i, template)
		},
	}
}
