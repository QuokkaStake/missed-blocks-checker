package telegram

import (
	"fmt"
	"main/pkg/utils"
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) ReplyRender(
	c tele.Context,
	templateName string,
	renderStruct interface{},
) error {
	template, err := reporter.TemplatesManager.Render(templateName, renderStruct)
	if err != nil {
		reporter.Logger.Error().Str("template", templateName).Err(err).Msg("Error rendering template")
		return c.Reply(fmt.Sprintf("Error rendering template: %s", err))
	}

	// to trim every string, mostly for tests
	templateSplit := strings.Split(template, "\n")
	templateTrimmed := utils.Map(templateSplit, strings.TrimSpace)
	templateJoined := strings.Join(templateTrimmed, "\n")

	return reporter.BotReply(c, templateJoined)
}
