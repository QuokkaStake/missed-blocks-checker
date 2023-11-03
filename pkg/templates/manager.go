package templates

import (
	"html/template"
	"main/pkg/constants"
	"main/pkg/types"
	"time"

	"github.com/rs/zerolog"
)

type Manager interface {
	Render(templateName string, data interface{}) (string, error)
	SerializeLink(link types.Link) template.HTML
	SerializeDate(date time.Time) string
	SerializeNotifiers(notifiers types.Notifiers) string
	SerializeNotifier(notifier *types.Notifier) string
	SerializeEvent(event types.RenderEventItem) string
}

func NewManager(logger zerolog.Logger, reporterType constants.ReporterName) Manager {
	switch reporterType {
	case constants.TelegramReporterName:
		return NewTelegramTemplateManager(logger)
	case constants.DiscordReporterName:
		return NewDiscordTemplateManager(logger)
	case constants.TestReporterName:
		fallthrough
	default:
		logger.Fatal().Str("type", string(reporterType)).Msg("Unknown reporter type")
		return nil
	}
}
