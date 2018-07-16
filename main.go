package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/svenwiltink/go-musicbot/bot"
	"github.com/svenwiltink/go-musicbot/bot/messageprovider/irc"
	"github.com/svenwiltink/go-musicbot/bot/messageprovider/terminal"
	"github.com/svenwiltink/go-musicbot/bot/messageprovider/rocketchat"
)

func main() {
	config, err := bot.LoadConfig("config.json")

	if err != nil {
		log.Printf("could not load config: %v", err)
		return
	}

	log.Println("loaded config")
	log.Println(config.MpvPath)

	if err != nil {
		log.Fatal(err)
	}

	messageProvider := chooseMessageProvider(config)
	err = messageProvider.Start()

	if err != nil {
		log.Fatal(err)
	}


	bot := bot.NewMusicBot(config, messageProvider)
	bot.Start()

	// Wait for a terminate signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Println("shutting down")
	bot.Stop()
}

func chooseMessageProvider(config *bot.Config) bot.MessageProvider {
	switch config.MessagePlugin {
	case "irc":
		log.Println("loading the irc message provider")
		return irc.New(config)
	case "terminal":
		log.Println("loading the terminal message provider")
		return terminal.New()
	case "rocketchat":
		log.Println("loading the rocketchat message provider")
		return rocketchat.New(config)
	default:
		log.Fatalf("unsupported message plugin: %s", config.MessagePlugin)
	}

	return nil
}
