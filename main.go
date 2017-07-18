package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/svenwiltink/go-musicbot/api"
	"github.com/svenwiltink/go-musicbot/bot"
	"github.com/svenwiltink/go-musicbot/config"
	"github.com/svenwiltink/go-musicbot/player"
	"github.com/svenwiltink/go-musicbot/songplayer"
	"github.com/svenwiltink/go-musicbot/util"
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
	if conf.LogLevel != "" {
		lvl, err := logrus.ParseLevel(conf.LogLevel)
		if err != nil {
			logrus.Errorf("main: Error reading loglevel [%s] %v", conf.LogLevel, err)
		} else {
			logrus.Infof("main: Setting loglevel to %s", lvl.String())
			logrus.SetLevel(lvl)
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
		logrus.Infof("main: Creating YoutubePlayer MPV control")
		ytMpvControl, err := util.NewMpvControl(conf.YoutubePlayer.MpvBinPath, conf.YoutubePlayer.MpvInputPath)
		if err != nil {
			logrus.Fatalf("main: Error creating YoutubePlayer MPV control: %v", err)
			musicBot.Announcef("%s[YoutubePlayer] Error creating MPV control: %v", bot.INVERSE_CHARACTER, err)
		}

		logrus.Infof("main: Creating YoutubePlayer")
		ytPlayer, err := songplayer.NewYoutubePlayer(conf.YoutubePlayer.YoutubeAPIKey, ytMpvControl)
		if err != nil {
			logrus.Errorf("main: Error creating YoutubePlayer: %v", err)
			musicBot.Announcef("%s[YoutubePlayer] Error creating player: %v", bot.INVERSE_CHARACTER, err)
		} else {
			playr.AddSongPlayer(ytPlayer)
		}
	}

	if conf.SpotifyPlayer.Enabled {
		logrus.Infof("main: Creating SpotifyPlayer")

		spPlayer, authURL, err := songplayer.NewSpotifyPlayer(conf.SpotifyPlayer.ClientID, conf.SpotifyPlayer.ClientSecret, conf.SpotifyPlayer.TokenFilePath, "", 0)
		if err != nil {
			logrus.Errorf("main: Error creating SpotifyPlayer: %v", err)
			musicBot.Announcef("%s[SpotifyPlayer] Error creating player: %v", bot.INVERSE_CHARACTER, err)
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
			musicBot.Announcef("[SpotifyPlayer] Authorisation: Add the external IP (%s) of the bot to your hosts file under 'musicbot' and visit:", ipStr)
			musicBot.Announce(authURL)
			spPlayer.AddAuthorisationListener(func() {
				playr.AddSongPlayer(spPlayer)
				musicBot.Announce("[SpotifyPlayer] The musicbot was successfully authorised!")
			})
		} else {
			playr.AddSongPlayer(spPlayer)
		}

		if conf.SpotifyPlayer.PlaybackDevice != "" {
			err = spPlayer.SetPlaybackDevice(conf.SpotifyPlayer.PlaybackDevice)
			if err != nil {
				logrus.Warnf("main: Error setting spotify playback device [%s] %v", conf.SpotifyPlayer.PlaybackDevice, err)
			} else {
				logrus.Infof("main: Successfully set spotify playback device [%s]", conf.SpotifyPlayer.PlaybackDevice)
			}
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
