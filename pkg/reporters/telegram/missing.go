package telegram

import (
	"fmt"
	"main/pkg/types"
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

	allValidators := reporter.Manager.GetValidators().ToSlice()
	activeValidators := utils.Filter(allValidators, func(v *types.Validator) bool {
		if !v.Active() {
			return false
		}

		signatureInfo := reporter.Manager.GetValidatorMissedBlocks(v)
		group, _ := reporter.Config.MissedBlocksGroups.GetGroup(signatureInfo.GetNotSigned())
		return group.Start > 0
	})

	if len(activeValidators) == 0 {
		return reporter.BotReply(c, "There are no missing validators!")
	}

	sort.Slice(activeValidators, func(firstIndex, secondIndex int) bool {
		first := activeValidators[firstIndex]
		second := activeValidators[secondIndex]
		firstSignature := reporter.Manager.GetValidatorMissedBlocks(first)
		secondSignature := reporter.Manager.GetValidatorMissedBlocks(second)

		return firstSignature.GetNotSigned() < secondSignature.GetNotSigned()
	})

	var sb strings.Builder

	for _, validator := range activeValidators {
		link := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
		signatureInfo := reporter.Manager.GetValidatorMissedBlocks(validator)
		group, _ := reporter.Config.MissedBlocksGroups.GetGroup(signatureInfo.GetNotSigned())

		sb.WriteString(fmt.Sprintf(
			"<strong>%s %s:</strong> %d missed blocks (%.2f%%)\n",
			group.EmojiEnd,
			reporter.SerializeLink(link),
			signatureInfo.GetNotSigned(),
			float64(signatureInfo.GetNotSigned())/float64(reporter.Config.BlocksWindow)*100,
		))
	}

	return reporter.BotReply(c, sb.String())
}
