package telegram

import (
	"fmt"
	"main/pkg/snapshot"
	"main/pkg/utils"
	"sort"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleListValidators(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got list validators query")

	validatorEntries := reporter.Manager.GetSnapshot().Entries.ToSlice()
	activeValidatorsEntries := utils.Filter(validatorEntries, func(v snapshot.Entry) bool {
		return v.Validator.Active()
	})

	if len(activeValidatorsEntries) == 0 {
		return reporter.BotReply(c, "There are no active validators!")
	}

	sort.Slice(activeValidatorsEntries, func(firstIndex, secondIndex int) bool {
		first := activeValidatorsEntries[firstIndex]
		second := activeValidatorsEntries[secondIndex]

		return first.SignatureInfo.GetNotSigned() < second.SignatureInfo.GetNotSigned()
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Validators' status on %s\n\n", reporter.Config.GetName()))

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
