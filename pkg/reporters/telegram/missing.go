package telegram

import (
	"fmt"
	"main/pkg/snapshot"
	"main/pkg/utils"
	"sort"
	"strings"

	tele "gopkg.in/telebot.v3"
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

	if len(activeValidatorsEntries) == 0 {
		return reporter.BotReply(c, "There are no missing validators!")
	}

	sort.Slice(activeValidatorsEntries, func(firstIndex, secondIndex int) bool {
		first := activeValidatorsEntries[firstIndex]
		second := activeValidatorsEntries[secondIndex]

		return first.SignatureInfo.GetNotSigned() < second.SignatureInfo.GetNotSigned()
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Validators missing blocks on %s\n\n", reporter.Config.GetName()))

	for _, validator := range activeValidatorsEntries {
		link := reporter.Config.ExplorerConfig.GetValidatorLink(validator.Validator)
		group, _ := reporter.Config.MissedBlocksGroups.GetGroup(validator.SignatureInfo.GetNotSigned())

		sb.WriteString(fmt.Sprintf(
			"<strong>%s %s:</strong> %d missed blocks (%.2f%%)\n",
			group.EmojiEnd,
			reporter.SerializeLink(link),
			validator.SignatureInfo.GetNotSigned(),
			float64(validator.SignatureInfo.GetNotSigned())/float64(reporter.Config.BlocksWindow)*100,
		))
	}

	return reporter.BotReply(c, sb.String())
}
