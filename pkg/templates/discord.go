package templates

import (
	"bytes"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"main/pkg/constants"
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
			return nil, errors.New("error converting template")
		} else {
			return convertedTemplate, nil
		}
	}

	allSerializers := map[string]any{
		"SerializeLink":             m.SerializeLink,
		"SerializeDate":             m.SerializeDate,
		"SerializeNotifier":         m.SerializeNotifier,
		"SerializeNotifiers":        m.SerializeNotifiers,
		"SerializeNotifiersNoLinks": m.SerializeNotifiersNoLinks,
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

func (m *DiscordTemplateManager) SerializeNotifiersNoLinks(notifiers types.Notifiers) string {
	notifiersNormalized := utils.Map(notifiers, func(n *types.Notifier) string {
		return "`@" + n.UserName + "`"
	})

	return strings.Join(notifiersNormalized, " ")
}

func (m *DiscordTemplateManager) SerializeNotifier(notifier *types.Notifier) string {
	return fmt.Sprintf("<@%s>", notifier.UserID)
}

func (m *DiscordTemplateManager) SerializeDate(date time.Time) string {
	return date.Format(time.RFC822)
}

func (m *DiscordTemplateManager) GetValidatorLink(validatorLink types.Link) htmlTemplate.HTML {
	if validatorLink.Href == "" {
		return htmlTemplate.HTML(validatorLink.Text)
	}

	return htmlTemplate.HTML(fmt.Sprintf(
		"%s (%s)",
		validatorLink.Text,
		m.SerializeLink(types.Link{
			Href: validatorLink.Href,
			Text: "link",
		}),
	))
}

func (m *DiscordTemplateManager) SerializeEvent(event types.RenderEventItem) string {
	renderData := types.ReportEventRenderData{
		Notifiers:     m.SerializeNotifiers(event.Notifiers),
		ValidatorLink: m.GetValidatorLink(event.ValidatorLink),
	}

	switch entry := event.Event.(type) {
	case events.ValidatorGroupChanged:
		if entry.IsIncreasing() {
			renderData.TimeToJail = fmt.Sprintf(" (%s till jail)", utils.FormatDuration(event.TimeToJail))
		}
	}

	return event.Event.Render(constants.FormatTypeMarkdown, renderData)
}
