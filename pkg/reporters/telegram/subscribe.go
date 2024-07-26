package telegram

import (
	"errors"
	"fmt"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) GetSubscribeCommand() Command {
	return Command{
		Name:    "subscribe",
		Execute: reporter.HandleSubscribe,
		MinArgs: 1,
		Usage:   "Usage: %s <validator address>",
	}
}

func (reporter *Reporter) HandleSubscribe(c tele.Context) (string, error) {
	username := c.Sender().Username
	if username == "" {
		username = c.Sender().FirstName
	} else {
		username = "@" + username
	}

	address := c.Args()[0]

	validator, found := reporter.Manager.GetValidator(address)
	if !found {
		return fmt.Sprintf(
			"Could not find a validator with address <code>%s</code> on %s",
			address,
			reporter.Config.GetName(),
		), errors.New("could not find validator")
	}

	added := reporter.Manager.AddNotifier(
		address,
		reporter.Name(),
		strconv.FormatInt(c.Sender().ID, 10),
		username,
	)

	if !added {
		return "You are already subscribed to this validator's notifications", errors.New("not subscribed")
	}

	validatorLink := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
	validatorLinkSerialized := reporter.TemplatesManager.SerializeLink(validatorLink)

	return fmt.Sprintf(
		"Subscribed to validator's notifications on %s: %s",
		reporter.Config.GetName(),
		validatorLinkSerialized,
	), nil
}
