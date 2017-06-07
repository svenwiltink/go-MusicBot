package config

import (
	"encoding/json"
	"os"
)

const (
	DEFAULT_LOGFILE_PATH = "musicbot.log"
	DEFAULT_QUEUE_PATH   = "queue.txt"
)

type MusicBot struct {
	LogFile   string
	QueuePath string

	IRC           IRC
	API           API
	YoutubePlayer YoutubePlayer
	SpotifyPlayer SpotifyPlayer
}

type IRC struct {
	Server        string
	Ssl           bool
	Channel       string
	Realname      string
	Nick          string
	Password      string
	Master        string
	WhiteListPath string
}

const DEFAULT_WHITELIST_PATH = "whitelist.txt"

type API struct {
	Host string
	Port int

	Username string
	Password string
}

const DEFAULT_API_PORT = 7070

type YoutubePlayer struct {
	Enabled      bool
	MpvBinPath   string
	MpvInputPath string

	YoutubeAPIKey string
}

const DEFAULT_MPV_BIN_PATH = "mpv"
const DEFAULT_MPV_INPUT_PATH = ".mpv-input"

type SpotifyPlayer struct {
	Enabled       bool
	Host          string
	UseConnect    bool
	ClientID      string
	ClientSecret  string
	TokenFilePath string
}

const DEFAULT_TOKEN_FILE_PATH = ".spotify-token"

func ReadConfig(path string) (conf *MusicBot, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(file)
	conf = &MusicBot{}
	conf.ApplyDefaults()

	err = decoder.Decode(&conf)
	if err != nil {
		return
	}

	return
}

func (c *MusicBot) ApplyDefaults() {
	c.QueuePath = DEFAULT_QUEUE_PATH
	c.IRC.WhiteListPath = DEFAULT_WHITELIST_PATH
	c.API.Port = DEFAULT_API_PORT
	c.SpotifyPlayer.Enabled = true
	c.SpotifyPlayer.TokenFilePath = DEFAULT_TOKEN_FILE_PATH
	c.YoutubePlayer.Enabled = true
	c.YoutubePlayer.MpvBinPath = DEFAULT_MPV_BIN_PATH
	c.YoutubePlayer.MpvInputPath = DEFAULT_MPV_INPUT_PATH
}
