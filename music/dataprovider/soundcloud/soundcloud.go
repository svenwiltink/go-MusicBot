package soundcloud

import (
	"github.com/svenwiltink/go-musicbot/music"
	"regexp"
)

var soundCloudRegex = regexp.MustCompile(`^https://soundcloud.com/([a-zA-Z0-9\-]+)/([a-zA-Z0-9\-]+)`)

type DataProvider struct {}

func (DataProvider) CanProvideData(song *music.Song) bool {
	return soundCloudRegex.MatchString(song.Path)
}

func (DataProvider) ProvideData(song *music.Song) error {
	matches := soundCloudRegex.FindStringSubmatch(song.Path)

	song.Artist = matches[1]
	song.Name = matches[2]

	return nil
}

