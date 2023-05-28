package telegram

import (
	"fmt"
	"main/pkg/constants"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleStatus(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got status query")

	reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "status")

	operatorAddresses := reporter.Manager.GetValidatorsForNotifier(reporter.Name(), c.Sender().Username)
	if len(operatorAddresses) == 0 {
		return reporter.BotReply(c, fmt.Sprintf(
			"You are not subscribed to any validator's notifications on %s.",
			reporter.Config.GetName(),
		))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"You are subscribed to the following validators' updates on %s:\n",
		reporter.Config.GetName(),
	))

	for _, operatorAddress := range operatorAddresses {
		validator, found := reporter.Manager.GetValidator(operatorAddress)
		if !found {
			return reporter.BotReply(c, fmt.Sprintf(
				"Could not find a validator with address <code>%s</code> on %s",
				operatorAddress,
				reporter.Config.GetName(),
			))
		}

		link := reporter.Config.ExplorerConfig.GetValidatorLink(validator)

		if validator.Jailed {
			sb.WriteString(fmt.Sprintf(
				"<strong>%s:</strong> jailed\n",
				reporter.SerializeLink(link),
			))
		} else if !validator.Active() {
			sb.WriteString(fmt.Sprintf(
				"<strong>%s:</strong> not in the active set\n",
				reporter.SerializeLink(link),
			))
		} else {
			signatureInfo := reporter.Manager.GetValidatorMissedBlocks(validator)
			sb.WriteString(fmt.Sprintf(
				"<strong>%s:</strong> %d missed blocks (%.2f%%)\n",
				reporter.SerializeLink(link),
				signatureInfo.GetNotSigned(),
				float64(signatureInfo.GetNotSigned())/float64(reporter.Config.BlocksWindow)*100,
			))
		}
	}

	return reporter.BotReply(c, sb.String())
}
