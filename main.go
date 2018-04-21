package main

import (
	"github.com/svenwiltink/go-musicbot/musicbot"
	"log"
	"github.com/svenwiltink/go-musicbot/musicbot/messageprovider/terminal"
)

func main() {
	config, err := musicbot.LoadConfig("config.json")

	log.Println("loaded config")

	if err != nil {
		log.Fatal(err)
	}

	//messageProvider := IrcMessageProvider.New(config)
	//messageProvider.Start()

	//messageProvider := DummyMessageProvider.New()
	//messageProvider.Start()

	messageProvider := terminal.New()
	messageProvider.Start()

	bot := musicbot.NewMusicBot(config, messageProvider)
	bot.Start()
}
