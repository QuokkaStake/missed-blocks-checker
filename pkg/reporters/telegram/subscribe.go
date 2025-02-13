package telegram

import (
	"fmt"
	"html"
	"main/pkg/constants"
	"strconv"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleSubscribe(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got subscribe query")

	reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "subscribe")

	username := c.Sender().Username
	if username == "" {
		username = c.Sender().FirstName
	} else {
		username = "@" + username
	}

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
			"Could not find a validator with address <code>%s</code> on %s!",
			address,
			reporter.Config.GetName(),
		))
	}

	added := reporter.Manager.AddNotifier(
		address,
		reporter.Name(),
		strconv.FormatInt(c.Sender().ID, 10),
		username,
	)

	if !added {
		return reporter.BotReply(c, "You are already subscribed to this validator's notifications!")
	}

	validatorLink := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
	validatorLinkSerialized := reporter.TemplatesManager.SerializeLink(validatorLink)

	return reporter.BotReply(c, fmt.Sprintf(
		"Subscribed to validator's notifications on %s: %s",
		reporter.Config.GetName(),
		validatorLinkSerialized,
	))
}
