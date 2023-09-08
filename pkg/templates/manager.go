package templates

import (
	"html/template"
	configPkg "main/pkg/config"
	"main/pkg/constants"
	reportPkg "main/pkg/report"
	statePkg "main/pkg/state"
	"main/pkg/types"
	"time"

	"github.com/rs/zerolog"
)

type Manager interface {
	Render(string, interface{}) (string, error)
	SerializeLink(link types.Link) template.HTML
	SerializeDate(date time.Time) string
	SerializeNotifiers(notifiers types.Notifiers) string
	SerializeNotifier(notifier *types.Notifier) string
	SerializeEntry(reportPkg.Entry, *statePkg.Manager, *configPkg.ChainConfig) string
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
