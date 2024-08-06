package commands

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

// updateDiscordRoles manages role synchronization between forum and Discord.
func updateDiscordRoles(s *discordgo.Session, m *discordgo.Member, GuildId string, forumGroups []string, rolesMap map[string]string) error {
	requiredRoles := make(map[string]struct{})
	for _, group := range forumGroups {
		if roleID, ok := rolesMap[group]; ok {
			requiredRoles[roleID] = struct{}{}
		}
	}

	var rolesToAdd []string
	var rolesToRemove []string
	currentRolesMap := make(map[string]struct{})
	for _, roleID := range m.Roles {
		currentRolesMap[roleID] = struct{}{}
	}

	for roleID := range requiredRoles {
		if _, hasRole := currentRolesMap[roleID]; !hasRole {
			rolesToAdd = append(rolesToAdd, roleID)
		}
	}

	for roleID := range currentRolesMap {
		if _, isRequired := requiredRoles[roleID]; !isRequired {
			rolesToRemove = append(rolesToRemove, roleID)
		}
	}

	// Update the user's roles on Discord.
	return updateRoles(s, GuildId, m.User.ID, rolesToAdd, rolesToRemove)
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
