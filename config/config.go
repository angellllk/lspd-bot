package config

import (
	"github.com/joho/godotenv"
	"log"
)

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
}
