package main

import (
	"fmt"
	"gitlab.transip.us/swiltink/go-MusicBot/api"
	"gitlab.transip.us/swiltink/go-MusicBot/bot"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"gitlab.transip.us/swiltink/go-MusicBot/playlist"
)

func main() {
	play := playlist.NewPlaylist()

	ytPlayer, err := player.NewYoutubePlayer()
	if err != nil {
		fmt.Printf("Error creating Youtube player: %v\n", err)
	} else {
		play.AddMusicPlayer(ytPlayer)
		fmt.Println("Added Youtube player")
	}

	spPlayer, err := player.NewSpotifyPlayer()
	if err != nil {
		fmt.Printf("Error creating Spotify player: %v\n", err)
	} else {
		play.AddMusicPlayer(spPlayer)
		fmt.Println("Added Spotify player")
	}

	// Initialize the API
	apiObject := api.NewAPI(play)
	go apiObject.Start()

	// Initialize the IRC bot
	botObject := bot.NewMusicBot(play)
	botObject.Start()
}
