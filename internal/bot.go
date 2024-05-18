package internal

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func RegisterCommands(s *discordgo.Session, guildID string) {
	_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, &discordgo.ApplicationCommand{
		Name:        "sync",
		Description: "Sync your roles from phpBB forum",
	})
	if err != nil {
		log.Fatal("cannot create slash command: ", err)
	}
}
