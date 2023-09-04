package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"main/pkg/constants"
)

func (reporter *Reporter) GetUnsubscribeCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "unsubscribe",
			Description: "Unsubscribe from validator's updates",
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
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.DiscordReporterName, "unsubscribe")

			options := i.ApplicationCommandData().Options
			address := options[0].Value.(string)

			user := i.User
			if user == nil {
				user = i.Member.User
			}
			if user == nil {
				reporter.BotRespond(s, i, "Could not fetch user!")
				return
			}

			validator, found := reporter.Manager.GetValidator(address)
			if !found {
				reporter.BotRespond(s, i, fmt.Sprintf(
					"Could not find a validator with address `%s` on %s!",
					address,
					reporter.Config.GetName(),
				))
				return
			}

			removed := reporter.Manager.RemoveNotifier(address, reporter.Name(), user.ID)

			if !removed {
				reporter.BotRespond(s, i, "You are not subscribed to this validator's notifications")
				return
			}

			validatorLink := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
			validatorLinkSerialized := reporter.SerializeLink(validatorLink)

			reporter.BotRespond(s, i, fmt.Sprintf(
				"Unsubscribed from validator's notifications on %s: %s",
				reporter.Config.GetName(),
				validatorLinkSerialized,
			))
		},
	}
}
