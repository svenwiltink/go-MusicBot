package main

import (
	"github.com/svenwiltink/go-musicbot/musicbot"
	"log"
	"github.com/svenwiltink/go-musicbot/musicbot/messageprovider/terminal"
	"github.com/svenwiltink/go-musicbot/musicbot/messageprovider/irc"
	"flag"
)

func main() {
	flag.Parse()
	config, err := musicbot.LoadConfig("config.json")

	log.Println("loaded config")

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
