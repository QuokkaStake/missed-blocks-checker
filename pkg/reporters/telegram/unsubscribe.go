package telegram

import (
	"errors"
	"fmt"
	"html"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) GetUnsubscribeCommand() Command {
	return Command{
		Name:    "unsubscribe",
		Execute: reporter.HandleUnsubscribe,
	}
}

func (reporter *Reporter) HandleUnsubscribe(c tele.Context) (string, error) {
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
