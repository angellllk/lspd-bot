package commands

import (
	"fmt"
	"github.com/angellllk/lspd-bot/internal"
	"github.com/angellllk/lspd-bot/internal/scraper"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
	"sync"
	"time"
)

// PhpBB represents the user role information fetched from the phpBB forum.
type PhpBB struct {
	Groups []string // A list of group IDs the user belongs to.
	Rank   string   // The rank of the user.
}

// CommandHandler manages interactions and commands from Discord users.
// It handles commands like /sync and manages the role synchronization
// between a phpBB forum and a Discord server.
type CommandHandler struct {
	// Session represents the active Discord session used for communication.
	Session *discordgo.Session

	// GuildID is the identifier for the Discord server where this handler operates.
	GuildID string

	// SyncChannelID is the identifier for the Discord sync channel.
	SyncChannelID string

	// Scraper is responsible for fetching data from the phpBB forum.
	Scraper *scraper.Scraper

	// RolesMap is the struct containing the file.json parsed that contains
	// roles that the Bot works with.
	RolesMap map[string]string
}

func (c *CommandHandler) MessageCreateHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if the message is in the allowed channel
	if m.ChannelID != c.SyncChannelID {
		return
	}

	// Add your bot's logic here
	if m.Content != "/sync" {
		s.ChannelMessageSend(m.ChannelID, "I can't read messages. Use slash (/) commands instead.")
	}
}

// InteractionCommandHandler is called when a Discord interaction is received.
// It dispatches the command to the appropriate handler function.
func (c *CommandHandler) InteractionCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Dispatch the command to the appropriate handler

	go c.DispatchCommand(i)
}

// DispatchCommand handles Discord interactions and determines
// if they should trigger specific bot actions. It processes interactions
// and dispatches them to appropriate command handlers.
func (c *CommandHandler) DispatchCommand(i *discordgo.InteractionCreate) {
	switch i.ApplicationCommandData().Name {
	case "sync":
		// Handle the /sync command
		go c.SyncCommandHandler(i)
	default:
		// Respond to unknown commands
		err := c.Session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unknown command. Please use /sync or other valid command.",
			},
		})
		if err != nil {
			log.Printf("Failed to respond to unknown command interaction: %v", err)
		}
	}
}

// SyncCommandHandler processes the /sync command from users.
// It synchronizes the user's roles between a phpBB forum and a Discord server
// based on their forum roles. It manages the synchronization flow and sends
// appropriate messages to the user throughout the process.
func (c *CommandHandler) SyncCommandHandler(i *discordgo.InteractionCreate) {
	startTime := time.Now()

	// Check if the command is /sync.
	if i.ApplicationCommandData().Name != "sync" {
		return
	}

	// Retrieve the channel where the command was issued.
	channel, err := c.Session.Channel(i.ChannelID)
	if err != nil {
		log.Fatal("Unable to fetch channel details: ", err)
	}

	// Ensure the command is issued in the "sync" channel.
	if channel.Name != "sync" {
		internal.SendDiscordResponse(c.Session, i.Interaction, fmt.Sprintf("This command can only be used in the <#%s> channel.", c.SyncChannelID))
		return
	}

	// Create a channel to communicate the results of the scraping process.
	ch := make(chan PhpBB, 1)
	defer close(ch)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		// Fetch user roles from the phpBB forum.
		groups, rank, errFetch := c.Scraper.FetchUserGroups(i.Member.Nick, i.Member.User.Username)
		if errFetch != nil {
			c.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: fmt.Sprintf("Unable to fetch roles from the forum. ||<@%s>||", i.Member.User.ID),
			})
			log.Printf("Error fetching phpBB roles: %v", errFetch)
			return
		}

		ch <- PhpBB{
			Groups: groups,
			Rank:   rank,
		}
	}()

	// Notify the user that the sync process has started.
	internal.SendDiscordResponse(c.Session, i.Interaction, "Syncing roles...")

	// Wait for the scraping process to complete.
	wg.Wait()

	result, ok := <-ch
	if !ok {
		c.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: "Failed to receive data from the processing routine.",
		})
		return
	}

	// Update the user's roles on Discord according to forum roles.
	errRoles := updateDiscordRoles(c.Session, i.Member, c.GuildID, result.Groups, c.RolesMap)
	if errRoles != nil {
		c.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Unable to update Discord roles. ||<@%s>||", i.Member.User.ID),
		})
		log.Printf("Error updating Discord roles: %v", errRoles)
		return
	}

	duration := time.Since(startTime)
	igName := strings.Split(i.Member.Nick, "/")
	// Confirm successful role synchronization.
	content := fmt.Sprintf(`# SYNC SUCCESSFUL
<@%s> Your roles have been synced. (**%s** %s)
Duration: %.3fs`, i.Member.User.ID, result.Rank, strings.TrimSpace(igName[0]), duration.Seconds())
	c.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
}
