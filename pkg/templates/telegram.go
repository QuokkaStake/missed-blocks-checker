package templates

import (
	"bytes"
	"fmt"
	"html/template"
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
		Templates: make(map[string]interface{}, 0),
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
