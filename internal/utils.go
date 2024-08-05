package internal

import "github.com/bwmarrin/discordgo"

func SendDiscordResponse(s *discordgo.Session, i *discordgo.Interaction, message string) {
	s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}
