package telegram

import (
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) GetSubscribeCommand() Command {
	return Command{
		Name:    "subscribe",
		Execute: reporter.HandleSubscribe,
	}
}

func (reporter *Reporter) HandleSubscribe(c tele.Context) (string, error) {
	username := c.Sender().Username
	if username == "" {
		username = c.Sender().FirstName
	} else {
		username = "@" + username
	}

	args := strings.Split(c.Text(), " ")
	if len(args) < 2 {
		return html.EscapeString(fmt.Sprintf(
			"Usage: %s <validator address>",
			args[0],
		)), errors.New("invalid invocation")
	}

	address := args[1]

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
