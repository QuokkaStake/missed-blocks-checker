package discord

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/utils"
	"sort"

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

			snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
			if !found {
				reporter.Logger.Info().
					Msg("No older snapshot on discord status query!")
				reporter.BotRespond(s, i, "Could not fetch user!")
				return
			}

			userEntries := snapshot.Entries.ByValidatorAddresses(operatorAddresses)

			entries := make([]statusEntry, len(userEntries))

			for index, entry := range userEntries {
				entries[index] = statusEntry{
					IsActive:  entry.IsActive,
					Validator: entry.Validator,
					Link:      reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator),
				}

				if entry.IsActive && !entry.Validator.Jailed {
					signatureInfo, err := reporter.Manager.GetValidatorMissedBlocks(entry.Validator)
					entries[index].Error = err
					entries[index].SigningInfo = signatureInfo
				}
			}

			sort.Slice(entries, func(i, j int) bool {
				first := entries[i]
				second := entries[j]

				if first.Validator.Jailed != second.Validator.Jailed {
					return utils.BoolToFloat64(second.Validator.Jailed)-utils.BoolToFloat64(first.Validator.Jailed) > 0
				}

				if first.IsActive != second.IsActive {
					return utils.BoolToFloat64(second.IsActive)-utils.BoolToFloat64(first.IsActive) > 0
				}

				return second.Validator.VotingPowerPercent < first.Validator.VotingPowerPercent
			})

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
