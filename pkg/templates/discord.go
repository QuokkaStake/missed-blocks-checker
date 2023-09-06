package templates

import (
	"bytes"
	"fmt"
	htmlTemplate "html/template"
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
