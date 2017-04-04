package main

import (
	"fmt"
	"gitlab.transip.us/swiltink/go-MusicBot/api"
	"gitlab.transip.us/swiltink/go-MusicBot/bot"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"gitlab.transip.us/swiltink/go-MusicBot/songplayer"
	"log"
)

func main() {
	conf, err := config.ReadConfig("conf.json")
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	playr := player.NewPlayer()

	ytPlayer, err := songplayer.NewYoutubePlayer()
	if err != nil {
		fmt.Printf("Error creating Youtube player: %v\n", err)
	} else {
		playr.AddSongPlayer(ytPlayer)
		fmt.Println("Added Youtube player")
	}

	spPlayer, err := songplayer.NewSpotifyPlayer()
	if err != nil {
		fmt.Printf("Error creating Spotify player: %v\n", err)
	} else {
		playr.AddSongPlayer(spPlayer)
		fmt.Println("Added Spotify player")
	}

	// Initialize the API
	apiObject := api.NewAPI(&conf.API, playr)
	go apiObject.Start()

	// Initialize the IRC bot
	botObject, err := bot.NewMusicBot(&conf.IRC, playr)
	if err != nil {
		fmt.Printf("Error creating IRC bot: %v\n", err)
		return
	}
	err = botObject.Start()
	if err != nil {
		fmt.Printf("Error starting IRC bot: %v\n", err)
		return
	}
}
