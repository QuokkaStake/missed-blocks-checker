package telegram

import (
	"fmt"
	tele "gopkg.in/telebot.v3"
	"main/pkg/snapshot"
	"main/pkg/utils"
	"sort"
)

func (reporter *Reporter) HandleMissingValidators(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got missing validators query")

	validatorEntries := reporter.Manager.GetSnapshot().Entries.ToSlice()
	activeValidatorsEntries := utils.Filter(validatorEntries, func(v snapshot.Entry) bool {
		if !v.Validator.Active() {
			return false
		}

		group, _ := reporter.Config.MissedBlocksGroups.GetGroup(v.SignatureInfo.GetNotSigned())
		return group.Start > 0
	})

	sort.Slice(activeValidatorsEntries, func(firstIndex, secondIndex int) bool {
		first := activeValidatorsEntries[firstIndex]
		second := activeValidatorsEntries[secondIndex]

		return first.SignatureInfo.GetNotSigned() < second.SignatureInfo.GetNotSigned()
	})

	render := missingValidatorsRender{
		Config: reporter.Config,
		Validators: utils.Map(activeValidatorsEntries, func(v snapshot.Entry) missingValidatorsEntry {
			link := reporter.Config.ExplorerConfig.GetValidatorLink(v.Validator)
			group, _ := reporter.Config.MissedBlocksGroups.GetGroup(v.SignatureInfo.GetNotSigned())
			link.Text = fmt.Sprintf("%s %s", group.EmojiEnd, v.Validator.Moniker)

			return missingValidatorsEntry{
				Validator:    v.Validator,
				Link:         link,
				NotSigned:    v.SignatureInfo.GetNotSigned(),
				BlocksWindow: reporter.Config.BlocksWindow,
			}
		}),
	}

	template, err := reporter.Render("Missing", render)
	if err != nil {
		return err
	}

	return reporter.BotReply(c, template)
}
