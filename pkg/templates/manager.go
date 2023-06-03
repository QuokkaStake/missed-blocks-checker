package templates

import (
	"fmt"
	"html/template"
	"main/pkg/constants"

	"github.com/rs/zerolog"
)

type Manager struct {
	Logger    zerolog.Logger
	Templates map[string]*template.Template
}

func NewManager(logger zerolog.Logger) *Manager {
	return &Manager{
		Logger:    logger.With().Str("component", "templates_manager").Logger(),
		Templates: make(map[string]*template.Template, 0),
	}
}

func (m *Manager) Render(
	templateName string,
	data interface{},
	formatType constants.FormatType,
) (string, error) {
	return m.RenderWithSerializers(templateName, data, formatType, make(map[string]any, 0))
}

func (m *Manager) RenderWithSerializers(
	templateName string,
	data interface{},
	formatType constants.FormatType,
	serializers map[string]any,
) (string, error) {
	switch formatType {
	case constants.FormatTypeHTML:
		return m.RenderHTML(templateName, data, serializers)
	default:
		return "", fmt.Errorf("unknown format type: %s", formatType)
	}
}
