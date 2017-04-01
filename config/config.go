package config

import (
	"encoding/json"
	"os"
)

type MusicBot struct {
	IRC IRC

	API API
}

type IRC struct {
	Server   string
	Ssl      bool
	Channel  string
	Realname string
	Nick     string
	Password string
	Master   string
}

const DEFAULT_API_PORT = 7070

type API struct {
	Host string
	Port int

	Username string
	Password string
}

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
	c.API.Port = DEFAULT_API_PORT
}
