package templates

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"main/pkg/events"
	"main/pkg/types"
	"main/pkg/utils"
	"main/templates"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type TelegramTemplateManager struct {
	Logger    zerolog.Logger
	Templates map[string]interface{}
}

func NewTelegramTemplateManager(logger zerolog.Logger) *TelegramTemplateManager {
	return &TelegramTemplateManager{
		Logger: logger.With().
			Str("component", "templates_manager").
			Str("reporter", "telegram").
			Logger(),
		Templates: make(map[string]interface{}),
	}
}

func (m *TelegramTemplateManager) GetHTMLTemplate(name string) (*template.Template, error) {
	if cachedTemplate, ok := m.Templates[name]; ok {
		m.Logger.Trace().Str("type", name).Msg("Using cached template")
		if convertedTemplate, ok := cachedTemplate.(*template.Template); !ok {
			return nil, fmt.Errorf("error converting template")
		} else {
			return convertedTemplate, nil
		}
	}

	m.Logger.Trace().Str("type", name).Msg("Loading template")

	allSerializers := map[string]any{
		"SerializeLink":      m.SerializeLink,
		"SerializeDate":      m.SerializeDate,
		"SerializeNotifier":  m.SerializeNotifier,
		"SerializeNotifiers": m.SerializeNotifiers,
	}

	t, err := template.New(name+".html").
		Funcs(allSerializers).
		ParseFS(templates.TemplatesFs, "telegram/"+name+".html")
	if err != nil {
		return nil, err
	}

	m.Templates[name] = t

	return t, nil
}

func (m *TelegramTemplateManager) Render(templateName string, data interface{}) (string, error) {
	templateToRender, err := m.GetHTMLTemplate(templateName)
	if err != nil {
		m.Logger.Error().Err(err).Str("type", templateName).Msg("Error loading template")
		return "", err
	}

	var buffer bytes.Buffer
	err = templateToRender.Execute(&buffer, data)
	if err != nil {
		m.Logger.Error().Err(err).Str("type", templateName).Msg("Error rendering template")
		return "", err
	}

	return buffer.String(), err
}

func (m *TelegramTemplateManager) SerializeDate(date time.Time) string {
	return date.Format(time.RFC822)
}

func (m *TelegramTemplateManager) SerializeLink(link types.Link) template.HTML {
	if link.Href == "" {
		return template.HTML(link.Text)
	}

	return template.HTML(fmt.Sprintf("<a href='%s'>%s</a>", link.Href, link.Text))
}

func (m *TelegramTemplateManager) SerializeNotifiers(notifiers types.Notifiers) string {
	notifiersNormalized := utils.Map(notifiers, m.SerializeNotifier)

	return strings.Join(notifiersNormalized, " ")
}

func (m *TelegramTemplateManager) SerializeNotifier(notifier *types.Notifier) string {
	if strings.HasPrefix(notifier.UserName, "@") {
		return notifier.UserName
	}

	return "@" + notifier.UserName
}

func (m *TelegramTemplateManager) SerializeEvent(event types.RenderEventItem) string {
	notifiersSerialized := " " + m.SerializeNotifiers(event.Notifiers)

	switch entry := event.Event.(type) {
	case events.ValidatorGroupChanged:
		timeToJailStr := ""

		if entry.IsIncreasing() {
			timeToJailStr = fmt.Sprintf(" (%s till jail)", utils.FormatDuration(event.TimeToJail))
		}

		return fmt.Sprintf(
			// a string like "🟡 <validator> is skipping blocks (> 1.0%)  (XXX till jail) <notifier> <notifier2>"
			"<strong>%s %s %s</strong>%s%s",
			entry.GetEmoji(),
			m.SerializeLink(event.ValidatorLink),
			html.EscapeString(entry.GetDescription()),
			timeToJailStr,
			notifiersSerialized,
		)
	case events.ValidatorJailed:
		return fmt.Sprintf(
			"<strong>❌ %s was jailed</strong>%s",
			m.SerializeLink(event.ValidatorLink),
			notifiersSerialized,
		)
	case events.ValidatorUnjailed:
		return fmt.Sprintf(
			"<strong>👌 %s was unjailed</strong>%s",
			m.SerializeLink(event.ValidatorLink),
			notifiersSerialized,
		)
	case events.ValidatorInactive:
		return fmt.Sprintf(
			"😔 <strong>%s is now not in the active set</strong>%s",
			m.SerializeLink(event.ValidatorLink),
			notifiersSerialized,
		)
	case events.ValidatorActive:
		return fmt.Sprintf(
			"✅ <strong>%s is now in the active set</strong>%s",
			m.SerializeLink(event.ValidatorLink),
			notifiersSerialized,
		)
	case events.ValidatorTombstoned:
		return fmt.Sprintf(
			"<strong>💀 %s was tombstoned</strong>%s",
			m.SerializeLink(event.ValidatorLink),
			notifiersSerialized,
		)
	case events.ValidatorCreated:
		return fmt.Sprintf(
			"<strong>💡New validator created: %s</strong>",
			m.SerializeLink(event.ValidatorLink),
		)
	default:
		return fmt.Sprintf("Unsupported event %+v\n", entry)
	}
}
