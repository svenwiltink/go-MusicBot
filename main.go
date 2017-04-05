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

	queueStorage := config.NewQueueStorage(conf.QueuePath)
	playr := player.NewPlayer()

	if conf.YoutubePlayer.Enabled {
		ytPlayer, err := songplayer.NewYoutubePlayer(conf.YoutubePlayer.MpvBinPath, conf.YoutubePlayer.MpvInputPath)
		if err != nil {
			fmt.Printf("Error creating Youtube player: %v\n", err)
		} else {
			playr.AddSongPlayer(ytPlayer)
			fmt.Println("Added Youtube player")
		}
	}

	if conf.SpotifyPlayer.Enabled {
		spPlayer, err := songplayer.NewSpotifyPlayer(conf.SpotifyPlayer.Host)
		if err != nil {
			fmt.Printf("Error creating Spotify player: %v\n", err)
		} else {
			playr.AddSongPlayer(spPlayer)
			fmt.Println("Added Spotify player")
		}
	}

	urls, err := queueStorage.ReadQueue()
	if err != nil {
		fmt.Printf("Error reading queue file: %v\n", err)
	} else {
		for _, url := range urls {
			playr.AddSongs(url)
		}
	}
	playr.AddListener("queue_updated", queueStorage.OnListUpdate)

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
