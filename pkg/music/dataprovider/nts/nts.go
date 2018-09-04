package nts

import (
	"regexp"
	"strings"

	"github.com/svenwiltink/go-musicbot/pkg/music"
)

var ntsRegex = regexp.MustCompile(`^nts[12]$`)

const (
	ntsStreamLink = "https://stream-relay-geo.ntslive.net/stream"
)

type DataProvider struct{}

func (DataProvider) CanProvideData(song *music.Song) bool {
	return ntsRegex.MatchString(song.Path)
}

func (DataProvider) ProvideData(song *music.Song) error {
	link := ntsStreamLink

	if strings.HasSuffix(song.Path, `2`) {
		link = link + `2`
	}

	song.Name = song.Path
	song.Artist = "nts"
	song.Path = link
	song.SongType = music.SongTypeStream

	return nil
}

func (DataProvider) Search(name string) ([]*music.Song, error) {
	return nil, nil
}
