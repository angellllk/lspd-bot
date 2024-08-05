package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	BotToken string
	Password string
	GuildID  string
}

func LoadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		BotToken: os.Getenv("DISCORD_BOT_TOKEN"),
		Password: os.Getenv("PASSWORD"),
		GuildID:  os.Getenv("GUILD_ID"),
	}

	if cfg.BotToken == "" || cfg.Password == "" || cfg.GuildID == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return cfg, nil
}
