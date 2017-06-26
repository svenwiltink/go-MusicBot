package player

import (
	"github.com/SvenWiltink/go-MusicBot/songplayer"
	"github.com/vansante/go-event-emitter"
	"time"
)

type Status int

const (
	PLAYING Status = 1 + iota
	PAUSED
	STOPPED
)

const (
	EVENT_QUEUE_ERROR_LOADING = "queue_error_loading"
	EVENT_QUEUE_LOADED        = "queue_loaded"
	EVENT_QUEUE_UPDATED       = "queue_updated"
	EVENT_STATS_ERROR_LOADING = "stats_error_loading"
	EVENT_STATS_LOADED        = "stats_loaded"
	EVENT_STATS_UPDATED       = "stats_updated"
	EVENT_ADDED_SONGS_USER    = "added_songs_user"
	EVENT_SONGS_ADDED         = "songs_added"
	EVENT_NEXT_SONG           = "next_song"
	EVENT_PREVIOUS_SONG       = "previous_song"
	EVENT_JUMP_SONG           = "jump_song"
	EVENT_PLAY_START          = "play_start"
	EVENT_PLAY_DONE           = "play_done"
	EVENT_SONG_SEEK           = "song_seek"
	EVENT_PAUSE               = "pause"
	EVENT_UNPAUSE             = "unpause"
	EVENT_STOP                = "stop"
	EVENT_QUEUE_DONE          = "queue_done"
)

type MusicPlayer interface {
	eventemitter.Observable

	GetSongPlayer(name string) (player songplayer.SongPlayer)
	GetSongPlayers() (players []songplayer.SongPlayer)
	GetHistory() (songs []songplayer.Playable)
	GetQueue() (songs []songplayer.Playable)
	GetCurrent() (song songplayer.Playable, remaining time.Duration)
	Add(url, actionUser string) (addedSongs []songplayer.Playable, err error)
	Insert(url string, position int, actionUser string) (addedSongs []songplayer.Playable, err error)
	ShuffleQueue()
	EmptyQueue()
	GetStatus() (status Status)
	Play() (song songplayer.Playable, err error)
	Seek(positionSeconds int) (err error)
	Next() (song songplayer.Playable, err error)
	Previous() (song songplayer.Playable, err error)
	Jump(deltaIndex int) (song songplayer.Playable, err error)
	Stop() (err error)
	Pause() (err error)
	GetStatistics() (stats *Statistics)
}
