package dummy

import (
	"log"
	"time"
	"sync"
	"github.com/svenwiltink/go-musicbot/musicplayer/musicprovider"
)

type SongPlayer struct {
	lock sync.Mutex
}

func (player *SongPlayer) CanPlay(song *musicprovider.Song) bool {
	return true
}

func (player *SongPlayer) Wait() {
	player.lock.Lock()
	defer player.lock.Unlock()

	time.Sleep(time.Second * 10)
}

func (player *SongPlayer) PlaySong(song *musicprovider.Song) error {
	player.lock.Lock()
	defer player.lock.Unlock()

	log.Printf("starting playback of %s", song.Name)
	return nil
}

func (player *SongPlayer) Play() error {
	player.lock.Lock()
	defer player.lock.Unlock()

	log.Printf("resuming playback")
	return nil
}

func (player *SongPlayer) Pause() error {
	player.lock.Lock()
	defer player.lock.Unlock()

	log.Printf("pausing playback")
	return nil
}

func NewSongPlayer() *SongPlayer {
	return &SongPlayer{}
}
