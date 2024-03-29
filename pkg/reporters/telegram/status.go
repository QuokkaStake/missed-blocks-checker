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

	entries := make([]statusEntry, len(operatorAddresses))

	for index, operatorAddress := range operatorAddresses {
		validator, found := reporter.Manager.GetValidator(operatorAddress)
		if !found {
			return reporter.BotReply(c, fmt.Sprintf(
				"Could not find a validator with address <code>%s</code> on %s",
				operatorAddress,
				reporter.Config.GetName(),
			))
		}

		entries[index] = statusEntry{
			Validator: validator,
			Link:      reporter.Config.ExplorerConfig.GetValidatorLink(validator),
		}

		if validator.Active() && !validator.Jailed {
			signatureInfo, err := reporter.Manager.GetValidatorMissedBlocks(validator)
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

		if first.Validator.Active() != second.Validator.Active() {
			return utils.BoolToFloat64(second.Validator.Active())-utils.BoolToFloat64(first.Validator.Active()) > 0
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
