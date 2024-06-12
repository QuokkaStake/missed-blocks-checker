package telegram

import (
	"fmt"
	"main/pkg/constants"
	"main/pkg/utils"
	"sort"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleStatus(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got status query")

	reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "status")

	operatorAddresses := reporter.Manager.GetValidatorsForNotifier(reporter.Name(), strconv.FormatInt(c.Sender().ID, 10))
	if len(operatorAddresses) == 0 {
		return reporter.BotReply(c, fmt.Sprintf(
			"You are not subscribed to any validator's notifications on %s.",
			reporter.Config.GetName(),
		))
	}

	snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
	if !found {
		reporter.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Msg("No older snapshot on telegram status query!")
		return reporter.BotReply(c, "Error getting your validators status")
	}

	userEntries := snapshot.Entries.ByValidatorAddresses(operatorAddresses)

	entries := make([]statusEntry, len(userEntries))

	for index, entry := range userEntries {
		entries[index] = statusEntry{
			Validator: entry.Validator,
			Link:      reporter.Config.ExplorerConfig.GetValidatorLink(entry.Validator),
			IsActive:  entry.IsActive,
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
		return err
	}

	return reporter.BotReply(c, template)
}
