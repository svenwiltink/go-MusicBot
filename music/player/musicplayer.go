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
	activeProvider music.Provider
	shouldStop     bool
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
	for !player.shouldStop {
		player.Status = StatusStopped

		song := player.Queue.WaitForNext()
		provider := player.getSuitablePlayer(song)
		player.activeProvider = provider
		err := provider.PlaySong(song)
		if err != nil {
			log.Println(err)
			continue
		}

		player.Status = StatusPlaying
		provider.Wait()

		log.Println("Song ended")
	}
}

func (player *MusicPlayer) Next() error {
	if player.Status == StatusPlaying {
		return player.activeProvider.Skip()
	}

	return fmt.Errorf("Nothing is playing")
}

func (player *MusicPlayer) Stop() {
	player.shouldStop = true
	for _, provider := range player.musicProviders {
		provider.Stop()
	}
}

// NewMusicPlayer creates a new MusicPlayer instance
func NewMusicPlayer(provider music.Provider) *MusicPlayer {
	instance := &MusicPlayer{
		Queue:          NewQueue(),
		musicProviders: make([]music.Provider, 0),
		shouldStop:     false,
	}

	instance.addMusicProvider(provider)
	return instance
}
