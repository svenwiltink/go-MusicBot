package player

import (
	"fmt"
	"log"

	"github.com/svenwiltink/go-musicbot/music"
)

// The possible statuses of the musicplayer
const (
	StatusPlaying = "playing"
	StatusPaused  = "pause"
	StatusStopped = "stopped"
)

// MusicPlayer is responsible for playing music
type MusicPlayer struct {
	Queue          *Queue
	Status         string
	musicProviders []music.Provider
}

func (player *MusicPlayer) addMusicProvider(provider music.Provider) {
	player.musicProviders = append(player.musicProviders, provider)
}

// AddSong tries to add the song to the Queue
func (player *MusicPlayer) AddSong(song *music.Song) error {

	suitablePlayer := player.getSuitablePlayer(song)
	if suitablePlayer == nil {
		return fmt.Errorf("no suitable player found for %+v", song)
	}

	player.Queue.append(song)
	return nil
}

func (player *MusicPlayer) getSuitablePlayer(song *music.Song) music.Provider {
	for _, provider := range player.musicProviders {
		if provider.CanPlay(song) {
			return provider
		}
	}

	return nil
}

// Start the MusicPlayer
func (player *MusicPlayer) Start() {
	log.Println("Starting music player")
	go player.playLoop()
}

func (player *MusicPlayer) playLoop() {
	for {
		song := player.Queue.WaitForNext()
		player := player.getSuitablePlayer(song)
		player.PlaySong(song)
		player.Wait()
		log.Println("Song ended")
	}
}

// NewMusicPlayer creates a new MusicPlayer instance
func NewMusicPlayer(provider music.Provider) *MusicPlayer {
	instance := &MusicPlayer{
		Queue:          NewQueue(),
		musicProviders: make([]music.Provider, 0),
	}

	instance.addMusicProvider(provider)
	return instance
}
