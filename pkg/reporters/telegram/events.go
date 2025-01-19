package telegram

import (
	"fmt"
	"html"
	"html/template"
	"main/pkg/constants"
	"main/pkg/types"
	"main/pkg/utils"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) HandleValidatorEventsList(c tele.Context) error {
	reporter.Logger.Info().
		Str("sender", c.Sender().Username).
		Str("text", c.Text()).
		Msg("Got validator events query")

	reporter.MetricsManager.LogReporterQuery(reporter.Config.Name, constants.TelegramReporterName, "events")

	args := strings.Split(c.Text(), " ")
	if len(args) < 2 {
		return reporter.BotReply(c, html.EscapeString(fmt.Sprintf(
			"Usage: %s <validator address>",
			args[0],
		)))
	}

	address := args[1]

	snapshot, found := reporter.SnapshotManager.GetNewerSnapshot()
	if !found {
		reporter.Logger.Info().
			Str("sender", c.Sender().Username).
			Str("text", c.Text()).
			Msg("No older snapshot on telegram events query!")
		return reporter.BotReply(c, "Error getting validator events!")
	}

	userEntries := snapshot.Entries.ByValidatorAddresses([]string{address})
	if len(userEntries) == 0 {
		return reporter.BotReply(c, "Validator is not found!")
	}

	jailsRaw, err := reporter.Manager.FindLastEventsByValidator(address)
	if err != nil {
		return reporter.BotReply(c, "Error searching for historical events!")
	}

	eventsRendered := utils.Map(jailsRaw, func(j types.HistoricalEvent) renderedHistoricalEvent {
		return renderedHistoricalEvent{
			Height: j.Height,
			Time:   j.Time,
			RenderedEvent: template.HTML(reporter.TemplatesManager.SerializeEvent(types.RenderEventItem{
				Event:         j.Event,
				Notifiers:     make(types.Notifiers, 0),
				ValidatorLink: reporter.Config.ExplorerConfig.GetValidatorLink(j.Event.GetValidator()),
			})),
		}
	})

	return reporter.ReplyRender(c, "Events", eventsRender{
		ValidatorLink: reporter.Config.ExplorerConfig.GetValidatorLink(userEntries[0].Validator),
		Events:        eventsRendered,
	})
}
