package telegram

import (
	"fmt"
	"main/pkg/constants"
	snapshotPkg "main/pkg/snapshot"
	"main/pkg/utils"
	"sort"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleMissingValidators(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got missing validators query")

	reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "missing")

	snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
	if !found {
		reporter.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Msg("No older snapshot on telegram validators query!")
		return reporter.BotReply(c, "Error getting validators list")
	}

	validatorEntries := snapshot.Entries.ToSlice()
	activeValidatorsEntries := utils.Filter(validatorEntries, func(v snapshotPkg.Entry) bool {
		if !v.IsActive {
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
		return err
	}

	return reporter.BotReply(c, template)
}
