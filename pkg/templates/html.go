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
)

func (m *Manager) GetHTMLTemplate(name string, serializers map[string]any) (*template.Template, error) {
	if cachedTemplate, ok := m.Templates[name]; ok {
		m.Logger.Trace().Str("type", name).Msg("Using cached template")
		return cachedTemplate, nil
	}

	m.Logger.Trace().Str("type", name).Msg("Loading template")

	allSerializers := map[string]any{
		"SerializeLink":      m.SerializeLink,
		"SerializeDate":      m.SerializeDate,
		"SerializeNotifier":  m.SerializeNotifier,
		"SerializeNotifiers": m.SerializeNotifiers,
	}

	for key, serializer := range serializers {
		allSerializers[key] = serializer
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

func (m *Manager) RenderHTML(
	templateName string,
	data interface{},
	serializers map[string]any,
) (string, error) {
	templateToRender, err := m.GetHTMLTemplate(templateName, serializers)
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

func (m *Manager) SerializeDate(date time.Time) string {
	return date.Format(time.RFC822)
}

func (m *Manager) SerializeLink(link types.Link) template.HTML {
	if link.Href == "" {
		return template.HTML(link.Text)
	}

	return template.HTML(fmt.Sprintf("<a href='%s'>%s</a>", link.Href, link.Text))
}

func (m *Manager) SerializeNotifiers(notifiers []string) string {
	notifiersNormalized := utils.Map(notifiers, m.SerializeNotifier)

	return strings.Join(notifiersNormalized, " ")
}

func (m *Manager) SerializeNotifier(notifier string) string {
	if strings.HasPrefix(notifier, "@") {
		return notifier
	}

	return "@" + notifier
}
