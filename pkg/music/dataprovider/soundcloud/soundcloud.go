package soundcloud

import (
	"regexp"

	"github.com/svenwiltink/go-musicbot/pkg/music"
	"github.com/svenwiltink/youtube-dl"
	"time"
)

var soundCloudRegex = regexp.MustCompile(`^https://soundcloud.com/([a-zA-Z0-9\-_]+)/([a-zA-Z0-9\-_]+)`)

type DataProvider struct{}

func (DataProvider) CanProvideData(song music.Song) bool {
	return soundCloudRegex.MatchString(song.Path)
}

func (DataProvider) ProvideData(song *music.Song) error {
	data, err := youtubedl.GetMetaData(song.Path)
	if err != nil {
		return err
	}

	song.Artist = data.Uploader
	song.Name = data.Title
	song.Duration = time.Second * time.Duration(data.Duration)

	return nil
}

func (DataProvider) Search(name string) ([]music.Song, error) {
	return nil, nil
}
