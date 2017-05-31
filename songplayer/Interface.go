package songplayer

import (
	"time"
)

type SongPlayer interface {
	Name() (name string)
	CanPlay(url string) (canPlay bool)
	GetSongs(url string) (songs []Playable, err error)
	Search(searchType SearchType, searchStr string, limit int) (songs []PlayableSearchResult, err error)
	Play(url string) (err error)
	Seek(positionSeconds int) (err error)
	Pause(pauseState bool) (err error)
	Stop() (err error)
}

type Playable interface {
	GetTitle() string
	GetDuration() time.Duration
	GetURL() string
	GetImageURL() string
}

type PlayableSearchResult interface {
	Playable

	GetType() SearchType
}
