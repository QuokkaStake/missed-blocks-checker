package telegram

import (
	"fmt"
	"main/pkg/constants"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleParams(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got params query")

	validators := reporter.Manager.GetValidators().ToSlice()
	entries := make([]notifierEntry, 0)

	for _, validator := range validators {
		link := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
		notifiers := reporter.Manager.GetNotifiersForReporter(validator.OperatorAddress, constants.TelegramReporterName)
		if len(notifiers) == 0 {
			continue
		}

		fmt.Printf("notifiers for %s: %+v\n", validator.OperatorAddress, notifiers)

		entries = append(entries, notifierEntry{
			Link:      link,
			Notifiers: notifiers,
		})
	}

	blockTime := reporter.Manager.GetBlockTime()
	maxTimeToJail := reporter.Manager.GetTimeTillJail(0)

	template, err := reporter.Render("Params", paramsRender{
		Config:        reporter.Config,
		BlockTime:     blockTime,
		MaxTimeToJail: maxTimeToJail,
	})
	if err != nil {
		return err
	}

	return reporter.BotReply(c, template)
}
