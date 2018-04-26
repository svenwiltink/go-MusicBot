package main

import (
	"log"

	"github.com/svenwiltink/go-musicbot/bot"
	"github.com/svenwiltink/go-musicbot/bot/messageprovider/irc"
	"github.com/svenwiltink/go-musicbot/bot/messageprovider/terminal"
)

func main() {
	config, err := bot.LoadConfig("config.json")

	log.Println("loaded config")
	log.Println(config.MpvPath)

	if err != nil {
		log.Fatal(err)
	}

	var messageProvider bot.MessageProvider
	switch config.MessagePlugin {
	case "irc":
		messageProvider = irc.New(config)
		log.Println("loaded the irc message provider")
		break
	case "terminal":
		messageProvider = terminal.New()
		log.Println("loaded the terminal message provider")
		break
	default:
		log.Fatalf("unsupported message plugin: %s", config.MessagePlugin)
	}

	messageProvider.Start()

	bot := bot.NewMusicBot(config, messageProvider)
	bot.Start()
}
