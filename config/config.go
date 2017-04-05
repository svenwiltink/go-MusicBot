package config

import (
	"encoding/json"
	"os"
)

const DEFAULT_QUEUE_PATH = "queue.txt"

type MusicBot struct {
	QueuePath string

	IRC IRC

	API API
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
}
