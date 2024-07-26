package telegram

import (
	"errors"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) GetParamsCommand() Command {
	return Command{
		Name:    "params",
		Execute: reporter.HandleParams,
	}
}

func (reporter *Reporter) HandleParams(c tele.Context) (string, error) {
	blockTime := reporter.Manager.GetBlockTime()
	maxTimeToJail := reporter.Manager.GetTimeTillJail(0)

	snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
	if !found {
		reporter.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Msg("No older snapshot on telegram params query!")
		return "", errors.New("no older snapshot")
	}

	activeValidators := snapshot.Entries.GetActive()
	var amount int
	if reporter.Config.IsConsumer.Bool {
		_, amount = snapshot.Entries.GetSoftOutOutThreshold(reporter.Config.ConsumerSoftOptOut)
	}

	return reporter.TemplatesManager.Render("Params", paramsRender{
		Config:                   reporter.Config,
		BlockTime:                blockTime,
		MaxTimeToJail:            maxTimeToJail,
		ConsumerOptOutValidators: amount,
		ValidatorsCount:          len(activeValidators),
	})
}
