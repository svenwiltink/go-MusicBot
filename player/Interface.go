package player

import (
	"github.com/vansante/go-event-emitter"
	"gitlab.transip.us/swiltink/go-MusicBot/songplayer"
	"time"
)

type Status int

const (
	PLAYING Status = 1 + iota
	PAUSED
	STOPPED
)

type MusicPlayer interface {
	eventemitter.Observable

	GetSongPlayer(name string) (player songplayer.SongPlayer)
	GetSongPlayers() (players []songplayer.SongPlayer)
	GetQueuedSongs() (songs []songplayer.Playable)
	GetCurrentSong() (song songplayer.Playable, remaining time.Duration)
	AddSongs(url string) (addedSongs []songplayer.Playable, err error)
	InsertSongs(url string, position int) (addedSongs []songplayer.Playable, err error)
	ShuffleQueue()
	EmptyQueue()
	GetStatus() (status Status)
	Play() (song songplayer.Playable, err error)
	Next() (song songplayer.Playable, err error)
	Stop() (err error)
	Pause() (err error)
}
