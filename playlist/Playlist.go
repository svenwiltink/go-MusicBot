package playlist

import (
	"errors"
	"fmt"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"math/rand"
	"time"
)

type MusicPlaylist struct {
	currentItem ItemInterface
	items       []ItemInterface
	status      Status

	currentPlayer player.MusicPlayerInterface
	players       []player.MusicPlayerInterface
}

func NewPlaylist() (playlist *MusicPlaylist) {
	playlist = &MusicPlaylist{
		status: STOPPED,
	}
	return
}

func (p *MusicPlaylist) AddMusicPlayer(player player.MusicPlayerInterface) {
	p.players = append(p.players, player)
}

func (p *MusicPlaylist) GetItems() (items []ItemInterface) {
	return p.items
}

func (p *MusicPlaylist) GetCurrentItem() (item ItemInterface) {
	return p.currentItem
}

func (p *MusicPlaylist) findPlayer(url string) (musicPlayer player.MusicPlayerInterface, err error) {
	for _, play := range p.players {
		if play.CanPlay(url) {
			musicPlayer = play
			return
		}
	}
	err = fmt.Errorf("No suitable musicplayer found to play %s", url)
	return
}

func (p *MusicPlaylist) AddItems(url string) (items []ItemInterface, err error) {
	musicPlayer, err := p.findPlayer(url)
	if err != nil {
		return
	}

	plItems, err := musicPlayer.GetItems(url)
	if err != nil {
		err = fmt.Errorf("[%s] Error getting items from url: %v", musicPlayer.Name(), err)
		return
	}

	for _, plItem := range plItems {
		p.items = append(p.items, &plItem)
	}
	return
}

func (p *MusicPlaylist) ShuffleList() {
	for i := range p.items {
		j := rand.Intn(i + 1)
		p.items[i], p.items[j] = p.items[j], p.items[i]
	}
}
func (p *MusicPlaylist) EmptyList() {
	p.items = make([]ItemInterface, 0)
}

func (p *MusicPlaylist) GetStatus() (status Status) {
	return p.status
}

func (p *MusicPlaylist) Start() (err error) {
	if p.status == STOPPED {
		_, err = p.Next()
	}
	return
}

func (p *MusicPlaylist) Next() (item ItemInterface, err error) {
	if len(p.items) == 0 {
		err = errors.New("Playlist is empty, no next available")
		return
	}
	if p.status == PLAYING || p.status == PAUSED {
		p.Stop()
	}

	item, p.items = p.items[0], p.items[1:]
	musicPlayer, err := p.findPlayer(item.URL())
	if err != nil {
		return
	}
	err = musicPlayer.Play(item.URL())
	if err != nil {
		err = fmt.Errorf("[%s] Error playing: %v", musicPlayer.Name(), err)
		return
	}
	p.currentPlayer = musicPlayer
	p.status = PLAYING

	go func() {
		time.Sleep(item.Duration() + time.Second)

		if len(p.items) > 0 {
			// TODO: Would prefer not to use recursion here, because we'll fill up the stack :')
			p.Next()
		}
	}()
	return
}

func (p *MusicPlaylist) Stop() (err error) {
	if p.status == STOPPED || p.currentPlayer == nil {
		err = errors.New("Nothing currently playing")
		return
	}
	err = p.currentPlayer.Stop()
	if err != nil {
		err = fmt.Errorf("[%s] Error stopping: %v", p.currentPlayer.Name(), err)
		return
	}
	p.currentPlayer = nil
	return
}

func (p *MusicPlaylist) Pause() (err error) {
	if p.status == STOPPED || p.currentPlayer == nil {
		err = errors.New("Nothing currently playing")
		return
	}

	err = p.currentPlayer.Pause(p.status != PAUSED)
	if err != nil {
		err = fmt.Errorf("[%s] Error (un)pausing [%v]: %v", p.currentPlayer.Name(), p.status != PAUSED, err)
		return
	}
	if p.status == PAUSED {
		p.status = PLAYING
	} else {
		p.status = PAUSED
	}
	return
}
