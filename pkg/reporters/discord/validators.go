package discord

import (
	"fmt"
	"main/pkg/constants"
	snapshotPkg "main/pkg/snapshot"
	"main/pkg/utils"
	"sort"

	"github.com/bwmarrin/discordgo"
)

func (reporter *Reporter) GetValidatorsCommand() *Command {
	return &Command{
		Info: &discordgo.ApplicationCommand{
			Name:        "validators",
			Description: "Get the list of all validators and their missing blocks",
		},
		Handler: func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "validators")

			snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
			if !found {
				reporter.Logger.Info().
					Msg("No older snapshot on discord validators query!")
				reporter.BotRespond(s, i, "Error getting validators list")
				return
			}

			validatorEntries := snapshot.Entries.ToSlice()
			activeValidatorsEntries := utils.Filter(validatorEntries, func(v snapshotPkg.Entry) bool {
				return v.Validator.Active()
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

			template, err := reporter.TemplatesManager.Render("Validators", render)
			if err != nil {
				reporter.Logger.Error().Err(err).Msg("Error rendering missing")
				return
			}

			reporter.BotRespond(s, i, template)
		},
	}
}
