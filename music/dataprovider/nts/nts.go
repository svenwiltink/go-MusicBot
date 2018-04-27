package nts

import (
	"github.com/svenwiltink/go-musicbot/music"
	"regexp"
	"strings"
)


var ntsRegex = regexp.MustCompile(`^nts[12]$`)

const (
	ntsStreamLink = "https://stream-relay-geo.ntslive.net/stream"
)

type DataProvider struct {}

func (DataProvider) CanProvideData(song *music.Song) bool {
	return ntsRegex.MatchString(song.Path)
}

func (DataProvider) ProvideData(song *music.Song) error {
	link := ntsStreamLink

	if strings.HasSuffix(song.Path, `2`) {
		link = link + `2`
	}

	song.Name = song.Path
	song.Artist = song.Path
	song.Path = link
	song.SongType = music.SongTypeStream

	return nil
}

