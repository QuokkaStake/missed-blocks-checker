package telegram

import (
	"main/pkg/constants"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleParams(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got params query")

	reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "params")

	blockTime := reporter.Manager.GetBlockTime()
	maxTimeToJail := reporter.Manager.GetTimeTillJail(0)

	snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
	if !found {
		reporter.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Msg("No older snapshot on telegram params query!")
		return reporter.BotReply(c, "Error getting params")
	}

	activeValidators := snapshot.Entries.GetActive()
	var amount int
	if reporter.Config.IsConsumer.Bool {
		_, amount = snapshot.Entries.GetSoftOutOutThreshold(reporter.Config.ConsumerSoftOptOut)
	}

	template, err := reporter.TemplatesManager.Render("Params", paramsRender{
		Config:                   reporter.Config,
		BlockTime:                blockTime,
		MaxTimeToJail:            maxTimeToJail,
		ConsumerOptOutValidators: amount,
		ValidatorsCount:          len(activeValidators),
	})
	if err != nil {
		return err
	}

	return reporter.BotReply(c, template)
}
