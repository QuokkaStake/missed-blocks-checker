package discord

import (
	"fmt"
	"main/pkg/constants"
	"strings"

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

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf(
				"You are subscribed to the following validators' updates on %s:\n",
				reporter.Config.GetName(),
			))

			for _, operatorAddress := range operatorAddresses {
				validator, found := reporter.Manager.GetValidator(operatorAddress)
				if !found {
					reporter.BotRespond(s, i, fmt.Sprintf(
						"Could not find a validator with address `%s` on %s",
						operatorAddress,
						reporter.Config.GetName(),
					))
					return
				}

				link := reporter.Config.ExplorerConfig.GetValidatorLink(validator)

				if validator.Jailed {
					sb.WriteString(fmt.Sprintf(
						"**%s:** jailed\n",
						reporter.SerializeLink(link),
					))
				} else if !validator.Active() {
					sb.WriteString(fmt.Sprintf(
						"**%s:** not in the active set\n",
						reporter.SerializeLink(link),
					))
				} else {
					if signatureInfo, err := reporter.Manager.GetValidatorMissedBlocks(validator); err != nil {
						sb.WriteString(fmt.Sprintf(
							"**%s:**: error getting validators missed blocks: %s",
							reporter.SerializeLink(link),
							err,
						))
					} else {
						sb.WriteString(fmt.Sprintf(
							"**%s:** %d missed blocks (%.2f%%)\n",
							reporter.SerializeLink(link),
							signatureInfo.GetNotSigned(),
							float64(signatureInfo.GetNotSigned())/float64(reporter.Config.BlocksWindow)*100,
						))
					}
				}
			}

			reporter.BotRespond(s, i, sb.String())
		},
	}
}
