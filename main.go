package main

import (
	"gitlab.transip.us/swiltink/go-MusicBot/api"
	"gitlab.transip.us/swiltink/go-MusicBot/bot"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
)

func main() {
	// Get a music player
	playerObject := player.NewMpvPlayer()

	// Initialize the IRC bot
	botObject := bot.NewMusicBot(playerObject)
	botObject.Start()

	// Initialize the API
	apiObject := api.NewAPI(playerObject)
	apiObject.Start()
}
