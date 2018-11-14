package m3u

import (
	"github.com/svenwiltink/go-musicbot/music"
	"regexp"
)

var m3uRegex = regexp.MustCompile(`\.m3u$`)

type DataProvider struct{}

func (DataProvider) CanProvideData(song *music.Song) bool {
	return m3uRegex.MatchString(song.Path)
}

func (DataProvider) ProvideData(song *music.Song) error {
	song.Artist = "m3u stream"
	song.Name = song.Path
	song.SongType = music.SongTypeStream

	return nil
}

func (DataProvider) Search(name string) ([]*music.Song, error) {
	return nil, nil
}
