package discord

import (
	"fmt"
	"main/pkg/constants"
	snapshotPkg "main/pkg/snapshot"
	"main/pkg/utils"
	"sort"

	"github.com/bwmarrin/discordgo"
)

func (reporter *Reporter) GetMissingCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "missing",
			Description: "Get the list of validators missing blocks",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "missing")

			snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
			if !found {
				reporter.Logger.Info().
					Msg("No older snapshot on discord missing query!")
				reporter.BotRespond(s, i, "Error getting validators list")
				return
			}

			validatorEntries := snapshot.Entries.ToSlice()
			activeValidatorsEntries := utils.Filter(validatorEntries, func(v snapshotPkg.Entry) bool {
				if !v.Validator.Active() {
					return false
				}

				group, _, _ := reporter.Config.MissedBlocksGroups.GetGroup(v.SignatureInfo.GetNotSigned())
				return group.Start > 0
			})

			sort.Slice(activeValidatorsEntries, func(firstIndex, secondIndex int) bool {
				first := activeValidatorsEntries[firstIndex]
				second := activeValidatorsEntries[secondIndex]

				return first.SignatureInfo.GetNotSigned() < second.SignatureInfo.GetNotSigned()
			})

			render := missingValidatorsRender{
				Config: reporter.Config,
				Validators: utils.Map(activeValidatorsEntries, func(v snapshotPkg.Entry) missingValidatorsEntry {
					link := reporter.Config.ExplorerConfig.GetValidatorLink(v.Validator)
					group, _, _ := reporter.Config.MissedBlocksGroups.GetGroup(v.SignatureInfo.GetNotSigned())
					link.Text = fmt.Sprintf("%s %s", group.EmojiEnd, v.Validator.Moniker)

					return missingValidatorsEntry{
						Validator:    v.Validator,
						Link:         link,
						NotSigned:    v.SignatureInfo.GetNotSigned(),
						BlocksWindow: reporter.Config.BlocksWindow,
					}
				}),
			}

			template, err := reporter.TemplatesManager.Render("Missing", render)
			if err != nil {
				reporter.Logger.Error().Err(err).Msg("Error rendering missing")
				return
			}

			chunks := utils.SplitStringIntoChunks(template, 2000)

			for _, chunk := range chunks {
				if err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: chunk,
					},
				}); err != nil {
					reporter.Logger.Error().Err(err).Msg("Error sending missing")
				}
			}
		},
	}
}
