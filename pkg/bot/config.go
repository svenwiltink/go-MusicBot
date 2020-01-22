package bot

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"time"
)

const (
	DefaultConfigFileLocation = "config.json"
	DefaultWhiteListFile      = "whitelist.txt"
	DefauultMaster            = "swiltink"
	DefaultCommandPrefix      = "!music"
)

type Config struct {
	WhiteListFile string           `json:"whitelistFile"`
	Master        string           `json:"master"`
	Irc           IRCConfig        `json:"irc"`
	Rocketchat    RocketchatConfig `json:"rocketchat"`
	Mattermost    MattermostConfig `json:"mattermost"`
	Youtube       YoutubeConfig    `json:"youtube"`
	MessagePlugin string           `json:"messageplugin"`
	CommandPrefix string           `json:"commandprefix"`
	MpvPath       string           `json:"mpvpath"`
	MpvSocket     string           `json:"mpvsocket"`
}

type IRCConfig struct {
	Server   string `json:"server"`
	Channel  string `json:"channel"`
	Nick     string `json:"nick"`
	RealName string `json:"realname"`
	Pass     string `json:"pass"`
	Ssl      bool   `json:"ssl"`
}

type RocketchatConfig struct {
	Server   string `json:"server"`
	Channel  string `json:"channel"`
	Username string `json:"username"`
	Pass     string `json:"pass"`
	Ssl      bool   `json:"ssl"`
}

type MattermostConfig struct {
	Server             string        `json:"server"`
	Teamname           string        `json:"teamname"`
	PrivateAccessToken string        `json:"privateAccessToken"`
	Channel            string        `json:"channel"`
	Ssl                bool          `json:"ssl"`
	ConnectionTimeout  time.Duration `json:"connectionTimeout"`
}

type YoutubeConfig struct {
	APIKey string `json:"apiKey"`
}

func (config *Config) applyDefaults() {
	config.WhiteListFile = DefaultWhiteListFile
	config.Master = DefauultMaster
	config.CommandPrefix = DefaultCommandPrefix
	config.Mattermost.ConnectionTimeout = 30
}

func (config *Config) CheckForErrors() error {
	if config.Mattermost.ConnectionTimeout <= 10 {
		return errors.Errorf("Mattermost ConnectionTimeout too low %d. Must be >= 10 seconds", config.Mattermost.ConnectionTimeout)
	}

	return nil
}

func LoadConfig(fileLocation string) (*Config, error) {
	file, err := os.Open(fileLocation)

	if err != nil {
		return nil, fmt.Errorf("unable to load config: %v", err)
	}

	config := &Config{}
	config.applyDefaults()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)

	if err != nil {
		return nil, fmt.Errorf("unable to decode config file: %v", err)
	}

	err = config.CheckForErrors()
	return config, err
}
