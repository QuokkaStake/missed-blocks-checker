package telegram

import (
	"main/pkg/constants"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleJailsCount(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got jails count query")

	reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "jailscount")

	snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
	if !found {
		reporter.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Msg("No older snapshot on telegram events query!")
		return reporter.BotReply(c, "Error getting validator events!")
	}

	jailsCount, err := reporter.Manager.FindAllJailsCount()
	if err != nil {
		return reporter.BotReply(c, "Error searching for jails count!")
	}

	jailsCountRendered := make([]renderedJailsCount, len(jailsCount))

	for index, validatorJailsCount := range jailsCount {
		validatorEntries := snapshot.Entries.ByValidatorAddresses([]string{validatorJailsCount.Validator})
		if len(validatorEntries) == 0 {
			return reporter.BotReply(c, "Validator is not found!")
		}

		jailsCountRendered[index] = renderedJailsCount{
			ValidatorLink: reporter.Config.ExplorerConfig.GetValidatorLink(validatorEntries[0].Validator),
			JailsCount:    validatorJailsCount.JailsCount,
		}
	}

	return reporter.ReplyRender(c, "JailsCount", jailsCountRendered)
}
