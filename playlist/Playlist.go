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

	players       []player.MusicPlayerInterface
	currentPlayer player.MusicPlayerInterface

	playTimer         *time.Timer
	endTime           time.Time
	remainingDuration time.Duration
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

func (p *MusicPlaylist) AddItems(url string) (addedItems []ItemInterface, err error) {
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
		addedItems = append(addedItems, &plItem)
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

func (p *MusicPlaylist) Play() (err error) {
	switch p.status {
	case STOPPED:
		_, err = p.Next()
	case PAUSED:
		err = p.Pause()
	}
	return
}

func (p *MusicPlaylist) playWait() {
	p.playTimer = time.NewTimer(p.endTime.Sub(time.Now()))

	// Wait for the timer to time out, or be canceled because of a STOP or something
	<-p.playTimer.C

	p.currentItem = nil
	if len(p.items) > 0 && p.status == PLAYING {
		p.Next()
	}
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
	musicPlayer, err := p.findPlayer(item.GetURL())
	if err != nil {
		return
	}
	err = musicPlayer.Play(item.GetURL())
	if err != nil {
		err = fmt.Errorf("[%s] Error playing: %v", musicPlayer.Name(), err)
		return
	}
	p.currentItem = item
	p.currentPlayer = musicPlayer
	p.status = PLAYING
	p.endTime = time.Now().Add(item.GetDuration())
	// Start waiting for the song to be done
	go p.playWait()
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
	p.currentItem = nil
	p.currentPlayer = nil
	if p.playTimer != nil {
		// Kill the current playWait()
		p.playTimer.Stop()
	}
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
		p.endTime = time.Now().Add(p.remainingDuration)
		// Restart the play wait goroutine with the new time
		go p.playWait()
	} else {
		p.status = PAUSED
		p.remainingDuration = p.endTime.Sub(time.Now())
		if p.playTimer != nil {
			// Kill the current playWait()
			p.playTimer.Stop()
		}
	}

	return
}
