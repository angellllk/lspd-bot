package commands

import (
	"fmt"
	"github.com/angellllk/lspd-bot/internal"
	"github.com/angellllk/lspd-bot/internal/scraper"
	"github.com/bwmarrin/discordgo"
	"log"
)

var GuildId string

func SyncCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name != "sync" {
		return
	}

	channel, err := s.Channel(i.ChannelID)
	if err != nil {
		log.Fatalf("can't get channel id: %v", err)
	}

	if channel.Name != "sync" {
		internal.SendDiscordResponse(s, i.Interaction, "You can use this only in #sync channel.")
		return
	}

	internal.SendDiscordResponse(s, i.Interaction, "Syncing roles...")

	phpbbRoles, errFetch := scraper.FetchUserGroups(i.Member.Nick, i.Member.User.Username)
	if errFetch != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Can't fetch roles from forums. ||<@%s>||", i.Member.User.ID),
		})
		log.Fatalf("couldn't fetch phpbb roles: %v", errFetch)
	}

	errRoles := updateDiscordRoles(s, i.Member, phpbbRoles)
	if errRoles != nil {
		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
			Content: fmt.Sprintf("Can't update Discord roles. ||<@%s>||", i.Member.User.ID),
		})
		log.Fatalf("couldn't update discord roles: %v", errRoles)
	}

	// Edit the original response to indicate that roles have been synced
	content := fmt.Sprintf("Roles synced. ||<@%s>||", i.Member.User.ID)
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
}

func updateDiscordRoles(s *discordgo.Session, m *discordgo.Member, forumRoles []string) (err error) {
	// Implement role synchronization logic here
	// This can involve checking current roles and updating them based on the roles fetched from phpBB

	var discordRoles []string
	roleIds := m.Roles

	allRolesIds, _ := s.GuildRoles(GuildId)

	for _, role := range allRolesIds {
		for _, memberRole := range roleIds {
			if role.ID == memberRole {
				discordRoles = append(discordRoles, role.Name)
				break
			}
		}
	}

	forum := make(map[string]bool, len(forumRoles))
	disc := make(map[string]bool, len(discordRoles))

	var add, del []string

	for _, v := range forumRoles {
		forum[v] = true
	}
	for _, v := range discordRoles {
		disc[v] = true
	}

	for k := range forum {
		if !disc[k] && k != "" {
			add = append(add, k)
		}
	}

	for k := range disc {
		if !forum[k] && k != "" {
			del = append(del, k)
		}
	}

	// Remove discord roles not found on the forums
	for _, roleId := range del {
		err = s.GuildMemberRoleRemove(GuildId, m.User.ID, roleId)
		if err != nil {
			return err
		}
	}

	var addDiscord []string
	for _, forumRole := range forumRoles {
		for _, role := range allRolesIds {
			if role.Name == forumRole {
				addDiscord = append(addDiscord, role.ID)
				break
			}
		}
	}

	// Add forum roles on discord
	for _, roleId := range addDiscord {
		err = s.GuildMemberRoleAdd(GuildId, m.User.ID, roleId)
		if err != nil {
			log.Printf("couldn't add role %s: %v \n", roleId, err)
		}
	}

	return err
}
