package templates

import (
	"bytes"
	"fmt"
	"main/pkg/types"
	"main/templates"
	"text/template"
)

func (m *Manager) GetMarkdownTemplate(
	name string,
	serializers map[string]any,
) (*template.Template, error) {
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

	for key, serializer := range serializers {
		allSerializers[key] = serializer
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

func (m *Manager) RenderMarkdown(
	templateName string,
	data interface{},
	serializers map[string]any,
) (string, error) {
	templateToRender, err := m.GetMarkdownTemplate(templateName, serializers)
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

func (m *Manager) SerializeMarkdownLink(link types.Link) string {
	if link.Href == "" {
		return link.Text
	}

	return fmt.Sprintf("[%s](%s)", link.Text, link.Href)
}
