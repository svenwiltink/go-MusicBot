package musicplayer

import (
	"github.com/svenwiltink/go-musicbot/musicplayer/musicprovider/dummy"
	"github.com/svenwiltink/go-musicbot/musicplayer/musicprovider"
	"fmt"
)

type MusicProvider interface {
	CanPlay(song *musicprovider.Song) bool
	PlaySong(song *musicprovider.Song) error
	Play() error
	Pause() error
	// wait for the current song to end
	Wait()
}

type MusicPlayer struct {
	Queue []*musicprovider.Song

	musicProviders []MusicProvider
}

func (player *MusicPlayer) addMusicProvider(provider MusicProvider) {
	player.musicProviders = append(player.musicProviders, provider)
}

func (player *MusicPlayer) AddSong(s string) error {
	song := &musicprovider.Song{
		Name: s,
		Path: s,
	}

	suitablePlayer := player.getSuitablePlayer(song)
	if suitablePlayer == nil {
		return fmt.Errorf("no suitable player found for %s", s)
	}

	return suitablePlayer.PlaySong(song)
}

func (player *MusicPlayer) getSuitablePlayer(song *musicprovider.Song) MusicProvider {
	for _, provider := range player.musicProviders {
		if provider.CanPlay(song) {
			return provider
		}
	}

	return nil
}

func NewMusicPlayer() *MusicPlayer {
	instance := &MusicPlayer{
		Queue:          make([]*musicprovider.Song, 0),
		musicProviders: make([]MusicProvider, 0),
	}

	instance.addMusicProvider(dummy.NewSongPlayer())
	return instance
}
