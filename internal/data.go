package internal

import (
	"github.com/bwmarrin/discordgo"
)

type Command interface {
	Definition(cmd string) *discordgo.ApplicationCommand
	OnCommand(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type Autocomplete interface {
	OnAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type ModalSubmit interface {
	CanHandle(customId string) bool
	OnModalSubmit(s *discordgo.Session, i *discordgo.InteractionCreate)
}

type MessageComponent interface {
	CanHandle(customId string) bool
	OnMessageComponent(s *discordgo.Session, i *discordgo.InteractionCreate)
}
