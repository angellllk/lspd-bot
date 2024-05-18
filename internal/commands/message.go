package commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Fatal("can't get channel ID")
	}

	if channel.Name != "sync" {
		return
	}

	if m.Content != "/sync" {
		_, errMsg := s.ChannelMessageSendReply(m.ChannelID, "Only /sync is available.", m.Reference())
		if errMsg != nil {
			log.Fatal("can't send message: ", errMsg)
		}
	}
}
