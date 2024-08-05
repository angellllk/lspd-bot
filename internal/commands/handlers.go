package commands

import (
	"fmt"
	"github.com/angellllk/lspd-bot/internal"
	"github.com/angellllk/lspd-bot/internal/scraper"
	"github.com/bwmarrin/discordgo"
	"log"
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

	// Scraper is responsible for fetching data from the phpBB forum.
	Scraper *scraper.Scraper
}

// SyncCommandHandler is called when a Discord interaction is received.
// It dispatches the command to the appropriate handler function.
func (c *CommandHandler) SyncCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
		go c.HandleSyncCommand(i)
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

// HandleSyncCommand processes the /sync command from users.
// It synchronizes the user's roles between a phpBB forum and a Discord server
// based on their forum roles. It manages the synchronization flow and sends
// appropriate messages to the user throughout the process.
func (c *CommandHandler) HandleSyncCommand(i *discordgo.InteractionCreate) {
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
		internal.SendDiscordResponse(c.Session, i.Interaction, "This command can only be used in the <#1238129021162885170> channel.")
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
	errRoles := updateDiscordRoles(c.Session, i.Member, c.GuildID, result.Groups)
	if errRoles != nil {
		c.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Unable to update Discord roles. ||<@%s>||", i.Member.User.ID),
		})
		log.Printf("Error updating Discord roles: %v", errRoles)
		return
	}

	duration := time.Since(startTime)
	// Confirm successful role synchronization.
	content := fmt.Sprintf(`# SYNC SUCCESSFUL
<@%s> Your roles have been synced. (**%s** %s)
Duration: %.3fs`, i.Member.User.ID, result.Rank, i.Member.Nick, duration.Seconds())
	c.Session.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
}
