package commands

import (
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
