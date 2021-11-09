package nts

import (
	"fmt"
	"github.com/svenwiltink/go-musicbot/pkg/music"
)

var streams = map[string]string {
	"nts1": "https://stream-relay-geo.ntslive.net/stream",
	"nts2": "https://stream-relay-geo.ntslive.net/stream2",
	"nts-feelings": "https://stream-mixtape-geo.ntslive.net/mixtape27",
	"nts-field": "https://stream-mixtape-geo.ntslive.net/mixtape23",
	"nts-memorylane": "https://stream-mixtape-geo.ntslive.net/mixtape6",
	"nts-4tothefloor": "https://stream-mixtape-geo.ntslive.net/mixtape5",
	"nts-thetube": "https://stream-mixtape-geo.ntslive.net/mixtape26",
	"nts-lowkey": "https://stream-mixtape-geo.ntslive.net/mixtape2",
	"nts-island": "https://stream-mixtape-geo.ntslive.net/mixtape21",
	"nts-raphouse": "https://stream-mixtape-geo.ntslive.net/mixtape22",
	"nts-sweat": "https://stream-mixtape-geo.ntslive.net/mixtape24",
	"nts-poolside": "https://stream-mixtape-geo.ntslive.net/mixtape4",
	"nts-slowfocus": "https://stream-mixtape-geo.ntslive.net/mixtape",
	"nts-expansions": "https://stream-mixtape-geo.ntslive.net/mixtape3",
}

type DataProvider struct{}

func (DataProvider) CanProvideData(song music.Song) bool {
	_, ok := streams[song.Path]
	return ok
}

func (DataProvider) ProvideData(song *music.Song) error {
	song.Name = song.Path
	song.Artist = "nts"
	song.Path = streams[song.Path]
	song.SongType = music.SongTypeStream
	return nil
}

func (DataProvider) Search(name string) ([]music.Song, error) {
	fmt.Println("trying to search NTS ", name)
	if name == "nts" {
		songs := make([]music.Song, 0, len(streams))

		for stream, _ := range streams {
			songs = append(songs, music.Song{
				Name:     stream,
				Artist:   "nts",
				SongType: music.SongTypeStream,
			})
		}

		return songs, nil
	}

	return nil, nil
}
