package commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

// updateDiscordRoles synchronizes user roles between Discord and a phpBB forum.
// It ensures that the user's roles on Discord match the roles fetched from the phpBB forum.
func updateDiscordRoles(s *discordgo.Session, m *discordgo.Member, GuildId string, forumRoles []string) (err error) {
	// Fetch the current roles of the user from Discord.
	var discordRoles []string
	roleIds := m.Roles

	// Retrieve all roles in the Discord guild.
	allRolesIds, _ := s.GuildRoles(GuildId)

	// Map user roles to their names.
	for _, role := range allRolesIds {
		for _, memberRole := range roleIds {
			if role.ID == memberRole {
				discordRoles = append(discordRoles, role.Name)
				break
			}
		}
	}

	// Create maps to track roles in both systems.
	forum := make(map[string]bool, len(forumRoles))
	disc := make(map[string]bool, len(discordRoles))

	var add, del []string

	// Populate forum roles map.
	for _, v := range forumRoles {
		forum[v] = true
	}

	// Populate Discord roles map.
	for _, v := range discordRoles {
		disc[v] = true
	}

	// Determine roles to add and remove.
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

	// Remove roles that are not found on the forums.
	for _, roleId := range del {
		err = s.GuildMemberRoleRemove(GuildId, m.User.ID, roleId)
		if err != nil {
			log.Printf("Error removing role %s: %v", roleId, err)
		}
	}

	var addDiscord []string
	// Map forum roles to Discord role IDs.
	for _, forumRole := range forumRoles {
		for _, role := range allRolesIds {
			if role.Name == forumRole {
				addDiscord = append(addDiscord, role.ID)
				break
			}
		}
	}

	// Add new roles on Discord.
	for _, roleId := range addDiscord {
		err = s.GuildMemberRoleAdd(GuildId, m.User.ID, roleId)
		if err != nil {
			log.Printf("Couldn't add role %s: %v", roleId, err)
		}
	}

	return nil
}
