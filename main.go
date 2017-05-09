package main

import (
	"fmt"
	"gitlab.transip.us/swiltink/go-MusicBot/api"
	"gitlab.transip.us/swiltink/go-MusicBot/bot"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"gitlab.transip.us/swiltink/go-MusicBot/songplayer"
	"gitlab.transip.us/swiltink/go-MusicBot/util"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	conf, err := config.ReadConfig("conf.json")
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	queueStorage := config.NewQueueStorage(conf.QueuePath)
	playr := player.NewPlayer()

	// Initialize the API
	apiObject := api.NewAPI(&conf.API, playr)
	go apiObject.Start()

	// Initialize the IRC bot
	musicBot, err := bot.NewMusicBot(&conf.IRC, playr)
	if err != nil {
		fmt.Printf("Error creating IRC bot: %v\n", err)
		return
	}
	err = musicBot.Start()
	if err != nil {
		fmt.Printf("Error starting IRC bot: %v\n", err)
		return
	}

	if conf.YoutubePlayer.Enabled {
		ytPlayer, err := songplayer.NewYoutubePlayer(conf.YoutubePlayer.MpvBinPath, conf.YoutubePlayer.MpvInputPath)
		if err != nil {
			fmt.Printf("Error creating Youtube player: %v\n", err)

			musicBot.Announce(fmt.Sprintf("[YoutubePlayer] Error creating player: %v", err))
		} else {
			playr.AddSongPlayer(ytPlayer)
			fmt.Println("Added Youtube player")
		}
	}

	if conf.SpotifyPlayer.Enabled && conf.SpotifyPlayer.UseConnect {
		spPlayer, authURL, err := songplayer.NewSpotifyConnectPlayer(conf.SpotifyPlayer.ClientID, conf.SpotifyPlayer.ClientSecret, "", 0)
		if err != nil {
			fmt.Printf("Error creating SpotifyConnect player: %v\n", err)

			musicBot.Announce(fmt.Sprintf("[SpotifyConnect] Error creating player: %v", err))
		} else {
			ips, err := util.GetExternalIPs()
			ipStr := "???"
			if err != nil {
				fmt.Printf("Error getting external IPs: %v\n", err)
			} else {
				ipStr = ""
				for _, ip := range ips {
					ipStr += ip.String() + " "
				}
				ipStr = strings.TrimSpace(ipStr)
			}
			musicBot.Announce(fmt.Sprintf("[SpotifyConnect] Authorisation: Please add the external IP (%s) of the musicbot to your host file under 'musicbot' and visit the follow URL: %s", ipStr, authURL))

			spPlayer.AddAuthorisationListener(func() {
				playr.AddSongPlayer(spPlayer)
				fmt.Println("Added SpotifyConnect player")

				musicBot.Announce("[SpotifyConnect] The musicbot was successfully authenticated!")
			})
		}
	} else if conf.SpotifyPlayer.Enabled && !conf.SpotifyPlayer.UseConnect {
		spPlayer, err := songplayer.NewSpotifyPlayer(conf.SpotifyPlayer.Host)
		if err != nil {
			fmt.Printf("Error creating Spotify player: %v\n", err)

			musicBot.Announce(fmt.Sprintf("[Spotify] Error creating player: %v", err))
		} else {
			playr.AddSongPlayer(spPlayer)
			fmt.Println("Added Spotify player")
		}
	}

	urls, err := queueStorage.ReadQueue()
	if err != nil {
		fmt.Printf("Error reading queue file: %v\n", err)

		musicBot.Announce(fmt.Sprintf("[Queue] Error loading queue: %v", err))
	} else {
		for _, url := range urls {
			playr.AddSongs(url)
		}
	}
	playr.AddListener("queue_updated", queueStorage.OnListUpdate)

	// Wait for a terminate signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	musicBot.Stop()
}
