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

type API struct {
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
	err = decoder.Decode(&conf)
	if err != nil {
		return
	}

	return
}
