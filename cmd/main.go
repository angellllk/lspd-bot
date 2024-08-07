package main

import (
	"github.com/angellllk/lspd-bot/config"
	"github.com/angellllk/lspd-bot/internal"
	"github.com/angellllk/lspd-bot/internal/commands"
	"github.com/angellllk/lspd-bot/internal/scraper"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	dg, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Fatal("error creating discord session: ", err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatal("error opening connection: ", err)
	}

	s := scraper.New(cfg.Password)
	ch := commands.CommandHandler{
		Session:       dg,
		GuildID:       cfg.GuildID,
		SyncChannelID: cfg.SyncChannelID,
		Scraper:       s,
		RolesMap:      cfg.RolesMap,
	}

	dg.AddHandler(ch.MessageCreateHandler)
	dg.AddHandler(ch.InteractionCommandHandler)

	internal.CleanupCommands(dg, cfg.GuildID)

	log.Println("Bot is running.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}
