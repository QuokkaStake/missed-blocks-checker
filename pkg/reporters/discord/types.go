package discord

import "github.com/bwmarrin/discordgo"

type Command struct {
	Info    *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}
