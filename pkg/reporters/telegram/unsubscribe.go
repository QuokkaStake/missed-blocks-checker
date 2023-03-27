package telegram

import (
	"fmt"
	"html"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleUnubscribe(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got unsubscribe query")

	args := strings.Split(c.Text(), " ")
	if len(args) < 2 {
		return reporter.BotReply(c, html.EscapeString(fmt.Sprintf(
			"Usage: %s <validator address>",
			args[0],
		)))
	}

	address := args[1]

	validator, found := reporter.Manager.GetValidator(address)
	if !found {
		return reporter.BotReply(c, fmt.Sprintf(
			"Could not find a validator with address <code>%s</code>",
			address,
		))
	}

	removed := reporter.Manager.RemoveNotifier(address, reporter.Name(), c.Sender().Username)

	if !removed {
		return reporter.BotReply(c, "You are not subscribed to this validator's notifications")
	}

	validatorLink := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
	validatorLinkSerialized := reporter.SerializeLink(validatorLink)

	return reporter.BotReply(c, fmt.Sprintf(
		"Unsubscribed from validator's notifications: %s",
		validatorLinkSerialized,
	))
}
