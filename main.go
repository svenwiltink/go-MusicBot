package main

import (
	"gitlab.transip.us/swiltink/go-MusicBot/bot"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
)

func main() {
	playerObject := player.NewMusicPlayer()
	botObject := bot.NewMusicBot(playerObject)
	botObject.Start()
}
