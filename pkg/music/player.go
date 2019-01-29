package music

import (
	"github.com/vansante/go-event-emitter"
	"time"
)

const (
	EventSongStarted    = "song-started"
	EventSongStartError = "song-start-error"
)

// Player is the wrapper around MusicProviders. This should keep track of the queue and control
// the MusicProviders
type Player interface {
	eventemitter.Observable
	Start()
	Search(string) ([]Song, error)
	SetVolume(percentage int) error
	IncreaseVolume(percentage int) (newVolume int, err error)
	DecreaseVolume(percentage int) (newVolume int, err error)
	GetVolume() (int, error)
	AddSong(song Song) error
	Next() error
	Pause() error
	Play() error
	Stop()
	GetStatus() PlayerStatus
	GetCurrentSong() (*Song, time.Duration)
	GetQueue() *Queue
}

type PlayerStatus string

const (
	PlayerStatusStarting PlayerStatus = "starting"
	PlayerStatusWaiting  PlayerStatus = "waiting"
	PlayerStatusLoading  PlayerStatus = "loading"
	PlayerStatusPlaying  PlayerStatus = "playing"
	PlayerStatusPaused   PlayerStatus = "paused"
)

func (s PlayerStatus) CanBeSkipped() bool {
	if s == PlayerStatusPlaying || s == PlayerStatusPaused {
		return true
	}

	return false
}
