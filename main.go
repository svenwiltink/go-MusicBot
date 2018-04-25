package main

import (
	"flag"
	"log"

	"github.com/svenwiltink/go-musicbot/musicbot"
	"github.com/svenwiltink/go-musicbot/musicbot/messageprovider/irc"
	"github.com/svenwiltink/go-musicbot/musicbot/messageprovider/terminal"
)

func main() {
	flag.Parse()
	config, err := musicbot.LoadConfig("config.json")

	log.Println("loaded config")
	log.Println(config.MpvPath)

	if err != nil {
		log.Fatal(err)
	}

	var messageProvider musicbot.MessageProvider
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

	bot := musicbot.NewMusicBot(config, messageProvider)
	bot.Start()
}
