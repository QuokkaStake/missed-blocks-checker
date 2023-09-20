package templates

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
	"main/pkg/events"
	"main/pkg/types"
	"main/pkg/utils"
	"main/templates"
	"strings"
	"text/template"
	"time"

	"github.com/rs/zerolog"
)

type DiscordTemplateManager struct {
	Logger    zerolog.Logger
	Templates map[string]interface{}
}

func NewDiscordTemplateManager(logger zerolog.Logger) *DiscordTemplateManager {
	return &DiscordTemplateManager{
		Logger: logger.With().
			Str("component", "templates_manager").
			Str("reporter", "discord").
			Logger(),
		Templates: make(map[string]interface{}, 0),
	}
}

func (m *DiscordTemplateManager) GetTemplate(name string) (*template.Template, error) {
	if cachedTemplate, ok := m.Templates[name]; ok {
		m.Logger.Trace().Str("type", name).Msg("Using cached template")
		if convertedTemplate, ok := cachedTemplate.(*template.Template); !ok {
			return nil, fmt.Errorf("error converting template")
		} else {
			return convertedTemplate, nil
		}
	}

	allSerializers := map[string]any{
		"SerializeLink":      m.SerializeLink,
		"SerializeDate":      m.SerializeDate,
		"SerializeNotifier":  m.SerializeNotifier,
		"SerializeNotifiers": m.SerializeNotifiers,
	}

	m.Logger.Trace().Str("type", name).Msg("Loading template")

	t, err := template.New(name+".md").
		Funcs(allSerializers).
		ParseFS(templates.TemplatesFs, "discord/"+name+".md")
	if err != nil {
		return nil, err
	}

	m.Templates[name] = t

	return t, nil
}

func (m *DiscordTemplateManager) Render(templateName string, data interface{}) (string, error) {
	templateToRender, err := m.GetTemplate(templateName)
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

func (m *DiscordTemplateManager) SerializeLink(link types.Link) htmlTemplate.HTML {
	if link.Href == "" {
		return htmlTemplate.HTML(link.Text)
	}

	// using <> to prevent auto-embed links, taken from here:
	// https://support.discord.com/hc/en-us/articles/206342858--How-do-I-disable-auto-embed-
	return htmlTemplate.HTML(fmt.Sprintf("[%s](<%s>)", link.Text, link.Href))
}

func (m *DiscordTemplateManager) SerializeNotifiers(notifiers types.Notifiers) string {
	notifiersNormalized := utils.Map(notifiers, m.SerializeNotifier)

	return strings.Join(notifiersNormalized, " ")
}

func (m *DiscordTemplateManager) SerializeNotifier(notifier *types.Notifier) string {
	return fmt.Sprintf("<@%s>", notifier.UserID)
}

func (m *DiscordTemplateManager) SerializeDate(date time.Time) string {
	return date.Format(time.RFC822)
}

func (m *DiscordTemplateManager) SerializeEvent(event types.RenderEventItem) string {
	notifiersSerialized := " " + m.SerializeNotifiers(event.Notifiers)

	validatorLink := fmt.Sprintf(
		"%s (%s)",
		event.ValidatorLink.Text,
		m.SerializeLink(types.Link{
			Href: event.ValidatorLink.Href,
			Text: event.ValidatorLink.Text,
		}),
	)

	switch entry := event.Event.(type) {
	case events.ValidatorGroupChanged:
		timeToJailStr := ""

		if entry.IsIncreasing() {
			timeToJailStr = fmt.Sprintf(" (%s till jail)", utils.FormatDuration(event.TimeToJail))
		}

		return fmt.Sprintf(
			// a string like "🟡 <validator> (link) is skipping blocks (> 1.0%)  (XXX till jail) <notifier> <notifier2>"
			"**%s %s %s**%s%s",
			entry.GetEmoji(),
			validatorLink,
			entry.GetDescription(),
			timeToJailStr,
			notifiersSerialized,
		)
	case events.ValidatorJailed:
		return fmt.Sprintf(
			"**❌ %s was jailed**%s",
			validatorLink,
			notifiersSerialized,
		)
	case events.ValidatorUnjailed:
		return fmt.Sprintf(
			"**👌 %s was unjailed**%s",
			validatorLink,
			notifiersSerialized,
		)
	case events.ValidatorInactive:
		return fmt.Sprintf(
			"😔 **%s is now not in the active set**%s",
			validatorLink,
			notifiersSerialized,
		)
	case events.ValidatorActive:
		return fmt.Sprintf(
			"✅ **%s is now in the active set**%s",
			validatorLink,
			notifiersSerialized,
		)
	case events.ValidatorTombstoned:
		return fmt.Sprintf(
			"**💀 %s was tombstoned**%s",
			validatorLink,
			notifiersSerialized,
		)
	case events.ValidatorCreated:
		return fmt.Sprintf(
			"**💡New validator created: %s**",
			validatorLink,
		)
	default:
		return fmt.Sprintf("Unsupported event %+v\n", entry)
	}
}
