package telegram

import (
	tele "gopkg.in/telebot.v3"
)

func (reporter *Reporter) GetHelpCommand() Command {
	return Command{
		Name:    "help",
		Execute: reporter.HandleHelp,
	}
}

func (reporter *Reporter) HandleHelp(c tele.Context) (string, error) {
	return reporter.TemplatesManager.Render("Help", reporter.Version)
}
