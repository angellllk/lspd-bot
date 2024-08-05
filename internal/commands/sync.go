package commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

// updateDiscordRoles manages role synchronization between forum and Discord.
func updateDiscordRoles(s *discordgo.Session, m *discordgo.Member, GuildId string, forumRoles []string) error {
	// Fetch the current roles of the user from Discord.
	discordRoles, err := fetchRoles(s, GuildId, m.Roles)
	if err != nil {
		return err
	}

	// Determine which roles need to be added or removed.
	add, del := determineRoleDifferences(forumRoles, discordRoles)

	// Update the user's roles on Discord.
	return updateRoles(s, GuildId, m.User.ID, add, del)
}

// fetchRoles retrieves the current roles for the user and maps them to their names.
func fetchRoles(s *discordgo.Session, GuildId string, roleIds []string) ([]string, error) {
	allRoles, err := s.GuildRoles(GuildId)
	if err != nil {
		return nil, err
	}

	roleNames := mapRolesToNames(allRoles, roleIds)
	return roleNames, nil
}

// mapRolesToNames maps role IDs to role names.
func mapRolesToNames(allRoles []*discordgo.Role, roleIds []string) []string {
	var roleNames []string
	roleIdSet := make(map[string]bool)

	for _, role := range allRoles {
		roleIdSet[role.ID] = true
	}

	for _, memberRole := range roleIds {
		if roleIdSet[memberRole] {
			roleNames = append(roleNames, findRoleNameById(allRoles, memberRole))
		}
	}

	return roleNames
}

// findRoleNameById finds the role name by its ID.
func findRoleNameById(allRoles []*discordgo.Role, roleId string) string {
	for _, role := range allRoles {
		if role.ID == roleId {
			return role.Name
		}
	}
	return ""
}

// determineRoleDifferences compares forum roles and Discord roles to determine which roles to add or remove.
func determineRoleDifferences(forumRoles, discordRoles []string) (add, del []string) {
	forumMap := make(map[string]bool)
	discordMap := make(map[string]bool)

	for _, role := range forumRoles {
		forumMap[role] = true
	}

	for _, role := range discordRoles {
		discordMap[role] = true
	}

	for role := range forumMap {
		if !discordMap[role] && role != "" {
			add = append(add, role)
		}
	}

	for role := range discordMap {
		if !forumMap[role] && role != "" {
			del = append(del, role)
		}
	}

	return
}

// updateRoles updates roles on Discord based on differences with forum roles.
func updateRoles(s *discordgo.Session, GuildId string, userId string, add, del []string) error {
	for _, roleId := range del {
		if err := s.GuildMemberRoleRemove(GuildId, userId, roleId); err != nil {
			log.Printf("Error removing role %s: %v", roleId, err)
		}
	}

	for _, roleId := range add {
		if err := s.GuildMemberRoleAdd(GuildId, userId, roleId); err != nil {
			log.Printf("Error adding role %s: %v", roleId, err)
		}
	}

	return nil
}
