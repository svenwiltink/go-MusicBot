package musicplayer

import (
	"fmt"
	"log"

	"github.com/svenwiltink/go-musicbot/musicplayer/musicprovider"
	"github.com/svenwiltink/go-musicbot/musicplayer/musicprovider/dummy"
)

type MusicProvider interface {
	CanPlay(song *musicprovider.Song) bool
	PlaySong(song *musicprovider.Song) error
	Play() error
	Pause() error
	// wait for the current song to end
	Wait()
}

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
	musicProviders []MusicProvider
}

func (player *MusicPlayer) addMusicProvider(provider MusicProvider) {
	player.musicProviders = append(player.musicProviders, provider)
}

// AddSong tries to add the song to the Queue
func (player *MusicPlayer) AddSong(s string) error {
	song := &musicprovider.Song{
		Name: s,
		Path: s,
	}

	suitablePlayer := player.getSuitablePlayer(song)
	if suitablePlayer == nil {
		return fmt.Errorf("no suitable player found for %s", s)
	}

	player.Queue.append(song)
	return nil
}

func (player *MusicPlayer) getSuitablePlayer(song *musicprovider.Song) MusicProvider {
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
func NewMusicPlayer() *MusicPlayer {
	instance := &MusicPlayer{
		Queue:          NewQueue(),
		musicProviders: make([]MusicProvider, 0),
	}

	instance.addMusicProvider(dummy.NewSongPlayer())
	return instance
}
