package main

import (
	"github.com/svenwiltink/go-musicbot/api"
	"github.com/svenwiltink/go-musicbot/bot"
	"github.com/svenwiltink/go-musicbot/config"
	"github.com/svenwiltink/go-musicbot/player"
	"github.com/svenwiltink/go-musicbot/songplayer"
	"github.com/svenwiltink/go-musicbot/util"
	"github.com/Sirupsen/logrus"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type LogFileHook struct {
	file      *os.File
	formatter logrus.Formatter
}

func (h *LogFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *LogFileHook) Fire(e *logrus.Entry) (err error) {
	buf, err := h.formatter.Format(e)
	if err != nil {
		return
	}
	_, err = h.file.Write(buf)
	return
}

func main() {
	// Set logrus to be the standard logger
	logger := logrus.New()
	logger.Formatter = &logrus.TextFormatter{
		DisableTimestamp: false,
		FullTimestamp:    true,
	}

	conf, err := config.ReadConfig("conf.json")
	if err != nil {
		logrus.Fatalf("main: Error reading musicbot config: %v", err)
		return
	}

	if conf.LogFile != "" {
		logFile, err := os.OpenFile(conf.LogFile, os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			logrus.Errorf("main: Error opening logfile [%s] %v", conf.LogFile, err)
		} else {
			defer logFile.Close()
			logrus.AddHook(&LogFileHook{
				file: logFile,
				formatter: &logrus.TextFormatter{
					DisableTimestamp: false,
					FullTimestamp:    true,
					DisableColors:    true,
				},
			})
		}
	}

	// Initialize the IRC bot
	musicBot, err := bot.NewMusicBot(conf)
	if err != nil {
		logrus.Fatalf("main: Error creating IRC MusicBot: %v", err)
		return
	}
	logrus.Infof("main: Starting IRC MusicBot")
	err = musicBot.Start()
	if err != nil {
		logrus.Fatalf("main: Error starting IRC MusicBot: %v", err)
		return
	}

	logrus.Infof("main: Creating new MusicPlayer with queuePath %s and statsPath %s", conf.QueuePath, conf.StatsPath)
	playr := player.NewPlayer(conf.QueuePath, conf.StatsPath)
	musicBot.SetPlayer(playr)

	// Initialize the API
	apiObject := api.NewAPI(&conf.API, playr)
	logrus.Infof("main: Starting HTTP API")
	go apiObject.Start()

	if conf.YoutubePlayer.Enabled {
		logrus.Infof("main: Creating YoutubePlayer")

		ytPlayer, err := songplayer.NewYoutubePlayer(conf.YoutubePlayer.YoutubeAPIKey, conf.YoutubePlayer.MpvBinPath, conf.YoutubePlayer.MpvInputPath)
		if err != nil {
			logrus.Errorf("main: Error creating YoutubePlayer: %v", err)
			musicBot.Announcef("[YoutubePlayer] Error creating player: %v", err)
		} else {
			playr.AddSongPlayer(ytPlayer)
		}
	}

	if conf.SpotifyPlayer.Enabled && conf.SpotifyPlayer.UseConnect {
		logrus.Infof("main: Creating SpotifyConnectPlayer")

		spPlayer, authURL, err := songplayer.NewSpotifyConnectPlayer(conf.SpotifyPlayer.ClientID, conf.SpotifyPlayer.ClientSecret, conf.SpotifyPlayer.TokenFilePath, "", 0)
		if err != nil {
			logrus.Errorf("main: Error creating SpotifyConnectPlayer: %v", err)
			musicBot.Announcef("[SpotifyConnectPlayer] Error creating player: %v", err)
		} else if authURL != "" {
			ips, err := util.GetExternalIPs()
			ipStr := "???"
			if err != nil {
				logrus.Warnf("main: Error getting external IPs: %v", err)
			} else {
				ipStr = ""
				for _, ip := range ips {
					ipStr += ip.String() + " "
				}
				ipStr = strings.TrimSpace(ipStr)
			}
			musicBot.Announcef("[SpotifyConnectPlayer] Authorisation: Add the external IP (%s) of the bot to your hosts file under 'musicbot' and visit:", ipStr)
			musicBot.Announce(authURL)
			spPlayer.AddAuthorisationListener(func() {
				playr.AddSongPlayer(spPlayer)
				musicBot.Announce("[SpotifyConnectPlayer] The musicbot was successfully authorised!")
			})
		} else {
			playr.AddSongPlayer(spPlayer)
		}
	}

	if conf.SpotifyPlayer.Enabled && !conf.SpotifyPlayer.UseConnect {
		logrus.Infof("main: Creating SpotifyPlayer")

		spPlayer, err := songplayer.NewSpotifyPlayer(conf.SpotifyPlayer.Host)
		if err != nil {
			logrus.Errorf("main: Error creating SpotifyPlayer: %v", err)
			musicBot.Announcef("[SpotifyPlayer] Error creating player: %v", err)
		} else {
			playr.AddSongPlayer(spPlayer)
		}
	}

	playr.Init()

	// Wait for a terminate signal
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	logrus.Infof("main: Shutting down")
	musicBot.Stop()
}
