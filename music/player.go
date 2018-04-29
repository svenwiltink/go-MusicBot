package music

import "github.com/vansante/go-event-emitter"

const (
	EventSongStarted = "song-started"
)

// Player is the wrapper around MusicProviders. This should keep track of the queue and control
// the MusicProviders
type Player interface {
	eventemitter.Observable
	Start()
	Search(string) ([]*Song, error)
	AddSong(song *Song) error
	Next() error
	Stop()
}
