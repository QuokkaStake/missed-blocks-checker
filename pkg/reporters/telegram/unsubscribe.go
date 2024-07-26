package telegram

import (
	"errors"
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) GetUnsubscribeCommand() Command {
	return Command{
		Name:    "unsubscribe",
		Execute: reporter.HandleUnsubscribe,
		MinArgs: 1,
		Usage:   "Usage: %s <validator address>",
	}
}

func (reporter *Reporter) HandleUnsubscribe(c tele.Context) (string, error) {
	address := c.Args()[0]

	validator, found := reporter.Manager.GetValidator(address)
	if !found {
		return fmt.Sprintf(
			"Could not find a validator with address <code>%s</code> on %s",
			address,
			reporter.Config.GetName(),
		), errors.New("could not find validator")
	}

	removed := reporter.Manager.RemoveNotifier(address, reporter.Name(), strconv.FormatInt(c.Sender().ID, 10))

	if !removed {
		return "You are not subscribed to this validator's notifications", errors.New("not subscribed")
	}

	validatorLink := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
	validatorLinkSerialized := reporter.TemplatesManager.SerializeLink(validatorLink)

	return fmt.Sprintf(
		"Unsubscribed from validator's notifications on %s: %s",
		reporter.Config.GetName(),
		validatorLinkSerialized,
	), nil
}
