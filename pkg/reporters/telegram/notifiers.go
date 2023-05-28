package telegram

import (
	"main/pkg/constants"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleNotifiers(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got notifiers query")

	reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "notifiers")

	validators := reporter.Manager.GetValidators().ToSlice()
	entries := make([]notifierEntry, 0)

	for _, validator := range validators {
		link := reporter.Config.ExplorerConfig.GetValidatorLink(validator)
		notifiers := reporter.Manager.GetNotifiersForReporter(validator.OperatorAddress, constants.TelegramReporterName)
		if len(notifiers) == 0 {
			continue
		}

		entries = append(entries, notifierEntry{
			Link:      link,
			Notifiers: notifiers,
		})
	}

	template, err := reporter.Render("Notifiers", notifierRender{
		Entries: entries,
		Config:  reporter.Config,
	})
	if err != nil {
		return err
	}

	return reporter.BotReply(c, template)
}
