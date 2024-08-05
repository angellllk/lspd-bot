package commands

import (
	"fmt"
	"github.com/angellllk/lspd-bot/internal"
	"github.com/angellllk/lspd-bot/internal/scraper"
	"github.com/bwmarrin/discordgo"
	"log"
)

// CommandHandler manages specific Discord operations
// such as handling messages and command interactions.
type CommandHandler struct {
	// Session represents the active Discord session.
	Session *discordgo.Session

	// GuildID is the identifier for the Discord server where this handler operates.
	GuildID string
}

// MessageHandler handles Discord messages and determines
// if they should trigger bot actions. Specifically, it
// processes messages in the "sync" channel and responds
// to commands with appropriate replies.
func (c *CommandHandler) MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from bots to prevent feedback loops.
	if m.Author.Bot {
		return
	}

	// Retrieve the channel details for the incoming message.
	channel, err := s.Channel(m.ChannelID)
	if err != nil {
		log.Fatal("Unable to fetch channel details: ", err)
	}

	// Only process messages from the "sync" channel.
	if channel.Name != "sync" {
		return
	}

	// Verify the message content is the expected command.
	if m.Content != "/sync" {
		_, errMsg := s.ChannelMessageSendReply(m.ChannelID, "Only /sync command is available.", m.Reference())
		if errMsg != nil {
			log.Fatal("Failed to send message: ", errMsg)
		}
	}
}

// SyncCommandHandler processes the /sync command from
// users, synchronizing their roles between a phpBB forum
// and a Discord server based on their forum roles.
// It manages the flow of syncing roles and responds to the
// user with appropriate messages throughout the process.
func (c *CommandHandler) SyncCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if the command is /sync.
	if i.ApplicationCommandData().Name != "sync" {
		return
	}

	// Retrieve the channel where the command was issued.
	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		log.Fatal("Unable to fetch channel details: ", err)
	}

	// Ensure the command is issued in the "sync" channel.
	if channel.Name != "sync" {
		internal.SendDiscordResponse(s, i.Interaction, "This command can only be used in the #sync channel.")
		return
	}

	// Notify the user that the sync process has started.
	internal.SendDiscordResponse(s, i.Interaction, "Syncing roles...")

	// Fetch user roles from the phpBB forum.
	phpbbRoles, errFetch := scraper.FetchUserGroups(i.Member.Nick, i.Member.User.Username)
	if errFetch != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Unable to fetch roles from the forum. ||<@%s>||", i.Member.User.ID),
		})
		log.Fatalf("Error fetching phpBB roles: %v", errFetch)
		return
	}

	// Update the user's roles on Discord according to forum roles.
	errRoles := updateDiscordRoles(s, i.Member, c.GuildID, phpbbRoles)
	if errRoles != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Unable to update Discord roles. ||<@%s>||", i.Member.User.ID),
		})
		log.Fatal("Error updating Discord roles: ", errRoles)
		return
	}

	// Confirm successful role synchronization.
	content := fmt.Sprintf("Roles successfully synced. ||<@%s>||", i.Member.User.ID)
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
}
