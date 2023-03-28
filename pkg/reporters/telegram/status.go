package telegram

import (
	"fmt"
	tele "gopkg.in/telebot.v3"
	"strings"
)

func (reporter *Reporter) HandleStatus(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got status query")

	operatorAddresses := reporter.Manager.GetValidatorsForNotifier(reporter.Name(), c.Sender().Username)
	if len(operatorAddresses) == 0 {
		return reporter.BotReply(c, "You are not subscribed to any validator's notifications.")
	}

	var sb strings.Builder
	sb.WriteString("You are subscribed to the following validators' updates:\n")

	for _, operatorAddress := range operatorAddresses {
		validator, found := reporter.Manager.GetValidator(operatorAddress)
		if !found {
			return reporter.BotReply(c, fmt.Sprintf(
				"Could not find a validator with address <code>%s</code>",
				operatorAddress,
			))
		}

		link := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
		signatureInfo := reporter.Manager.GetValidatorMissedBlocks(validator)

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
			sb.WriteString(fmt.Sprintf(
				"<strong>%s:</strong> %d missed blocks (%.2f%%)\n",
				reporter.SerializeLink(link),
				signatureInfo.GetNotSigned(),
				float64(signatureInfo.GetNotSigned())/float64(reporter.Config.ChainConfig.BlocksWindow)*100,
			))
		}
	}

	return reporter.BotReply(c, sb.String())
}
