package main

import (
	"gitlab.transip.us/swiltink/go-MusicBot/api"
	"gitlab.transip.us/swiltink/go-MusicBot/bot"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
)

func main() {
	// Get a music player
	playerObject := player.NewMpvPlayer()

	// Initialize the API
	apiObject := api.NewAPI(playerObject)
	go apiObject.Start()

	// Initialize the IRC bot
	botObject := bot.NewMusicBot(playerObject)
	botObject.Start()
}
