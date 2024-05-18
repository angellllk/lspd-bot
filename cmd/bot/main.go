package main

import (
	"github.com/angellllk/lspd-bot/config"
	"github.com/angellllk/lspd-bot/internal"
	"github.com/angellllk/lspd-bot/internal/commands"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.LoadConfig()

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if len(token) == 0 {
		log.Fatal("no bot token provided")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("error creating discord session: ", err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatal("error opening connection: ", err)
	}

	dg.AddHandler(commands.MessageHandler)
	dg.AddHandler(commands.SyncCommandHandler)

	internal.RegisterCommands(dg, "")

	log.Println("Bot is running.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}
