package player

import (
	"fmt"
	"log"

	"github.com/svenwiltink/go-musicbot/pkg/music"

	"errors"
	"time"

	eventemitter "github.com/vansante/go-event-emitter"
)

// MusicPlayer is responsible for playing music
type MusicPlayer struct {
	*eventemitter.Emitter
	Queue           *music.Queue
	Status          music.PlayerStatus
	dataProviders   []music.DataProvider
	musicProviders  []music.Provider
	activeProvider  music.Provider
	currentSong     *music.Song
	shouldStop      bool
	currentSongEnds time.Time
}

func (player *MusicPlayer) GetQueue() *music.Queue {
	return player.Queue
}

func (player *MusicPlayer) Pause() error {
	if player.Status != music.PlayerStatusPlaying {
		return errors.New("cannot pause, nothing is playing")
	}

	err := player.activeProvider.Pause()

	if err == nil {
		player.Status = music.PlayerStatusPaused
	}

	return err

}

func (player *MusicPlayer) Play() error {
	if player.Status != music.PlayerStatusPaused {
		return errors.New("cannot resume, music is not paused")
	}

	err := player.activeProvider.Play()

	if err == nil {
		player.Status = music.PlayerStatusPlaying
	}

	return err
}

func (player *MusicPlayer) GetStatus() music.PlayerStatus {
	return player.Status
}

func (player *MusicPlayer) GetCurrentSong() (*music.Song, time.Duration) {
	if player.currentSong != nil {
		return player.currentSong, time.Until(player.currentSongEnds).Round(time.Second)
	}

	return nil, time.Duration(0)
}

func (player *MusicPlayer) SetVolume(percentage int) error {
	for _, provider := range player.musicProviders {
		if err := provider.SetVolume(percentage); err != nil {
			return err
		}
	}

	return nil
}

func (player *MusicPlayer) IncreaseVolume(percentage int) (newVolume int, err error) {
	for _, provider := range player.musicProviders {
		newVolume, err = provider.GetVolume()
		if err != nil {
			return 0, err
		}

		newVolume = newVolume + percentage

		if newVolume > 100 {
			newVolume = 100
		}

		if newVolume < 0 {
			newVolume = 0
		}

		err = provider.SetVolume(newVolume)
		if err != nil {
			return 0, err
		}
	}

	return newVolume, nil
}

func (player *MusicPlayer) DecreaseVolume(percentage int) (newVolume int, err error) {
	return player.IncreaseVolume(-percentage)
}

func (player *MusicPlayer) GetVolume() (int, error) {
	if player.activeProvider != nil {
		return player.activeProvider.GetVolume()
	}

	return 0, errors.New("nothing is playing")
}

func (player *MusicPlayer) Search(searchString string) ([]music.Song, error) {
	songs := make([]music.Song, 0)

	for _, provider := range player.dataProviders {
		results, err := provider.Search(searchString)
		if err != nil {
			return nil, err
		}

		if results != nil {
			songs = append(songs, results...)
		}
	}

	return songs, nil
}

// AddSong tries to add the song to the Queue
func (player *MusicPlayer) AddSong(song music.Song) (music.Song, error) {
	// assume it is a song unless the dataprovider changes it to a stream
	song.SongType = music.SongTypeSong

	dataProvider := player.getSuitableDataProvider(song)

	if dataProvider == nil {
		return song, fmt.Errorf("no dataprovider found for %+v", song)
	}

	err := dataProvider.ProvideData(&song)

	log.Printf("provided song data: %+v", song)

	if err != nil {
		return song, fmt.Errorf("could not get data for song: %v", err)
	}

	suitablePlayer := player.getSuitablePlayer(song)

	if suitablePlayer == nil {
		return song, fmt.Errorf("no suitable player found for %+v", song)
	}

	player.Queue.Append(song)
	return song, nil
}

func (player *MusicPlayer) getSuitableDataProvider(song music.Song) music.DataProvider {
	for _, provider := range player.dataProviders {
		if provider.CanProvideData(song) {
			return provider
		}
	}

	return nil
}

func (player *MusicPlayer) getSuitablePlayer(song music.Song) music.Provider {
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
		player.Status = music.PlayerStatusWaiting
		log.Println("Waiting for song")
		song := player.Queue.WaitForNext()
		player.currentSong = &song

		provider := player.getSuitablePlayer(song)
		player.activeProvider = provider

		player.Status = music.PlayerStatusLoading
		err := provider.PlaySong(song)

		if err != nil {
			log.Println(err)
			player.EmitEvent(music.EventSongStartError, song, err)
			continue
		}

		player.currentSongEnds = time.Now().Add(song.Duration)
		player.EmitEvent(music.EventSongStarted, song)
		player.Status = music.PlayerStatusPlaying
		provider.Wait()

		log.Println("Song ended")
	}
}

func (player *MusicPlayer) Next() error {
	fmt.Printf("current player status: %v", player.Status)

	if !player.Status.CanBeSkipped() {
		return fmt.Errorf("nothing is playing")
	}

	err := player.activeProvider.Skip()
	if err != nil {
		return err
	}

	if player.Status == music.PlayerStatusPaused {
		err = player.activeProvider.Play()
		return err
	}

	return nil
}

func (player *MusicPlayer) Stop() {
	player.shouldStop = true
	for _, provider := range player.musicProviders {
		provider.Stop()
	}
}

func (player *MusicPlayer) AddPlaylist(playlistUrl string) (*music.Playlist, error) {
	playlist := music.Playlist{}

	for _, provider := range player.dataProviders {
		result, err := provider.AddPlaylist(playlistUrl)
		if err != nil {
			return nil, err
		}

		if result != nil {
			playlist.Title = result.Title
			playlist.Songs = result.Songs
		}
	}

	for _, song := range playlist.Songs {
		_, err := player.AddSong(song)

		if err != nil {
			return nil, err
		}
	}

	return &playlist, nil
}

// NewMusicPlayer creates a new MusicPlayer instance
func NewMusicPlayer(providers []music.Provider, dataProviders []music.DataProvider) *MusicPlayer {
	instance := &MusicPlayer{
		Emitter:        eventemitter.NewEmitter(false),
		Queue:          music.NewQueue(),
		musicProviders: providers,
		dataProviders:  dataProviders,
		shouldStop:     false,
	}

	return instance
}
