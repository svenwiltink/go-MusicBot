package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

const (
	DEFAULT_LOGFILE_PATH = "musicbot.log"
	DEFAULT_LOGLEVEL     = "info"
	DEFAULT_QUEUE_PATH   = "queue.txt"
	DEFAULT_STATS_PATH   = "musicbot-stats.json"
)

func GetDefaultOSConfigPath() (path string) {
	path = "/etc/go-musicbot/config.json"

	//android darwin dragonfly freebsd linux nacl netbsd openbsd plan9 solaris windows zos
	goos := runtime.GOOS
	switch goos {
	case "darwin":
		path = filepath.Join(os.Getenv("HOME") + "Library/Application Support/go-musicbot/config.json")
	case "freebsd":
		path = "/usr/local/etc/go-musicbot/config.json"
	case "win":
		path = filepath.Join(os.Getenv("programdata") + "go-musicbot/config.json")
	}
	return
}

type MusicBot struct {
	LogFile   string
	LogLevel  string
	QueuePath string
	StatsPath string

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
	Enabled bool

	MpvBinPath    string
	MpvInputPath  string
	YoutubeAPIKey string
}

const (
	DEFAULT_YOUTUBE_MPV_BIN_PATH   = "mpv"
	DEFAULT_YOUTUBE_MPV_INPUT_PATH = ".yt-mpv-input"
)

type SpotifyPlayer struct {
	Enabled        bool
	ClientID       string
	ClientSecret   string
	TokenFilePath  string
	PlaybackDevice string
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
	c.LogFile = DEFAULT_LOGFILE_PATH
	c.LogLevel = DEFAULT_LOGLEVEL
	c.QueuePath = DEFAULT_QUEUE_PATH
	c.StatsPath = DEFAULT_STATS_PATH

	c.IRC.WhiteListPath = DEFAULT_WHITELIST_PATH

	c.API.Port = DEFAULT_API_PORT

	c.SpotifyPlayer.Enabled = true
	c.SpotifyPlayer.TokenFilePath = DEFAULT_TOKEN_FILE_PATH

	c.YoutubePlayer.Enabled = true
	c.YoutubePlayer.MpvBinPath = DEFAULT_YOUTUBE_MPV_BIN_PATH
	c.YoutubePlayer.MpvInputPath = DEFAULT_YOUTUBE_MPV_INPUT_PATH
}
