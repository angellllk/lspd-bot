package commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func updateDiscordRoles(s *discordgo.Session, m *discordgo.Member, GuildId string, forumRoles []string) (err error) {
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
			log.Printf("error removing role %s: %v", roleId, err)
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

	return nil
}
