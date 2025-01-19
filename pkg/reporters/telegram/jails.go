package telegram

import (
	"html/template"
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/utils"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleJailsList(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got jails list query")

	reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "jails")

	jailsRaw, err := reporter.Manager.FindLastEventsByType([]constants.EventName{
		constants.EventValidatorJailed,
		constants.EventValidatorTombstoned,
	})
	if err != nil {
		return reporter.BotReply(c, "Error searching for historical events!")
	}

	jailsRendered := utils.Map(jailsRaw, func(j types.HistoricalEvent) jailsEntry {
		return jailsEntry{
			Height: j.Height,
			Time:   j.Time,
			RenderedEvent: template.HTML(reporter.TemplatesManager.SerializeEvent(types.RenderEventItem{
				Event:         j.Event,
				Notifiers:     make(types.Notifiers, 0),
				ValidatorLink: reporter.Config.ExplorerConfig.GetValidatorLink(j.Event.GetValidator()),
			})),
		}
	})

	return reporter.ReplyRender(c, "Jails", jailsRendered)
}
