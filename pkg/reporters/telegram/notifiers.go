package telegram

import (
	"main/pkg/constants"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) GetNotifiersCommand() Command {
	return Command{
		Name:    "validators",
		Execute: reporter.HandleNotifiers,
	}
}

func (reporter *Reporter) HandleNotifiers(c tele.Context) (string, error) {
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

	return reporter.TemplatesManager.Render("Notifiers", notifierRender{
		Entries: entries,
		Config:  reporter.Config,
	})
}
