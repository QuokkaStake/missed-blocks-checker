package telegram

import (
	"fmt"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleSubscribe(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got subscribe query")

	args := strings.Split(c.Text(), " ")
	if len(args) < 2 {
		return reporter.BotReply(c, fmt.Sprintf("Usage: %s <validator address", args[0]))
	}

	address := args[1]

	added := reporter.Manager.AddNotifier(address, reporter.Name(), c.Sender().Username)

	if !added {
		return reporter.BotReply(c, "You are already subscribed to this validator's notifications")
	}

	return reporter.BotReply(c, fmt.Sprintf(
		"Subscribed to validator notifications: %s",
		address,
	))
}
