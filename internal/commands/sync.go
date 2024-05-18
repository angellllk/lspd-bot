package commands

import (
	"github.com/angellllk/lspd-bot/internal/scraper"
	"github.com/bwmarrin/discordgo"
	"log"
)

func SyncCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name != "sync" {
		return
	}

	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Error: Could not retrieve channel information.",
			},
		})
		return
	}

	if channel.Name != "sync" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "You can use this only in the #sync channel",
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Syncing your roles...",
		},
	})

	phpbbRoles, errFetch := scraper.FetchUserGroups(i.Member.Nick, i.Member.User.Username)
	if errFetch != nil {
		_, errMsg := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error: Can't fetch user groups",
		})
		if errMsg != nil {
			log.Fatal("can't send followup message: ", errMsg)
		}
		return
	}

	errRoles := updateDiscordRoles(s, i.Member, phpbbRoles)
	if errRoles != nil {
		_, errMsg := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Error: Could not sync roles.",
		})
		if errMsg != nil {
			log.Fatal("can't send followup message: ", errMsg)
		}
		return
	}

	_, errMsg := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: "Roles synced successfully!",
	})
	if errMsg != nil {
		log.Fatal("can't send followup message: ", errMsg)
	}
}

func updateDiscordRoles(s *discordgo.Session, m *discordgo.Member, roles []string) error {
	// Implement role synchronization logic here
	// This can involve checking current roles and updating them based on the roles fetched from phpBB
	return nil
}
