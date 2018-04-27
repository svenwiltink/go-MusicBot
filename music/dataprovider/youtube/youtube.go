package youtube

import (
	"github.com/svenwiltink/go-musicbot/music"
	"regexp"
)

const (
	youTubeVideoURL    = "https://www.youtube.com/watch?v=%s"
	youTubePlaylistURL = "https://www.youtube.com/watch?v=%s&list=%s"

	MaxYoutubeItems = 500
)

var youtubeURLRegex = regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com|youtu\.?be)/.+$`)

type DataProvider struct {

}

func (provider *DataProvider) CanProvideData(song *music.Song) bool {
	return youtubeURLRegex.MatchString(song.Path)
}

func (provider *DataProvider) ProvideData(song *music.Song) error {
	song.Name = song.Path
}

