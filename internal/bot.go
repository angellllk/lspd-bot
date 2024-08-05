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

func CleanupCommands(s *discordgo.Session, guildID string) {
	cmds, err := s.ApplicationCommands(s.State.User.ID, guildID)
	if err != nil {
		log.Fatal("couldn't list commands: ", err)
	}

	for _, cmd := range cmds {
		err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
		if err != nil {
			log.Printf("couldn't delete command %s: %v", cmd.Name, err)
		}
	}
}
