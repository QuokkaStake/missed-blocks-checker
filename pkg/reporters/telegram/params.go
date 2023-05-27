package telegram

import (
	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleParams(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got params query")

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
