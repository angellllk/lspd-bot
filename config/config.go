package config

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	BotToken      string
	Password      string
	GuildID       string
	SyncChannelID string
	RolesMap      map[string]string
}

func LoadConfig(path ...string) (*Config, error) {
	var err error
	if len(path) > 0 {
		err = godotenv.Load(path[0])
		if err != nil {
			return nil, err
		}
	} else {
		err = godotenv.Load()
		if err != nil {
			return nil, err
		}
	}

	jsonName := os.Getenv("ROLES_JSON")
	jsonB, errRead := os.ReadFile(jsonName)
	if errRead != nil {
		return nil, errRead
	}

	var roles map[string]string
	errParse := json.Unmarshal(jsonB, &roles)

	if errParse != nil {
		return nil, errParse
	}

	cfg := &Config{
		BotToken:      os.Getenv("DISCORD_BOT_TOKEN"),
		Password:      os.Getenv("PASSWORD"),
		GuildID:       os.Getenv("GUILD_ID"),
		SyncChannelID: os.Getenv("SYNC_CHANNEL_ID"),
		RolesMap:      roles,
	}

	if cfg.BotToken == "" || cfg.Password == "" || cfg.GuildID == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return cfg, nil
}
