package main

import (
	"github.com/svenwiltink/go-musicbot/musicbot"
	IrcMessageProvider "github.com/svenwiltink/go-musicbot/musicbot/messageprovider/irc"
	"log"
)

func main() {
	config, err := musicbot.LoadConfig("config.json")

	log.Printf("loaded config: %+v", config)

	if err != nil {
		log.Fatal(err)
	}

	messageProvider := IrcMessageProvider.New(config)
	messageProvider.Start()

	bot := musicbot.NewMusicBot(config, messageProvider)
	bot.Start()
}
