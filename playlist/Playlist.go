package playlist

import (
	"errors"
	"fmt"
	"github.com/vansante/go-event-emitter"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type MusicPlaylist struct {
	*eventemitter.Emitter

	currentItem ItemInterface
	items       []ItemInterface
	status      Status

	players       []player.MusicPlayerInterface
	currentPlayer player.MusicPlayerInterface

	playTimer         *time.Timer
	endTime           time.Time
	remainingDuration time.Duration

	controlMutex sync.Mutex
}

func NewPlaylist() (playlist *MusicPlaylist) {
	playlist = &MusicPlaylist{
		Emitter: eventemitter.NewEmitter(),
		status:  STOPPED,
	}
	return
}

func (p *MusicPlaylist) GetPlayer(name string) (plyr player.MusicPlayerInterface) {
	for _, plr := range p.players {
		if strings.ToLower(plr.Name()) == strings.ToLower(name) {
			plyr = plr
			return
		}
	}
	return
}

func (p *MusicPlaylist) GetPlayers() (players []player.MusicPlayerInterface) {
	return p.players
}

func (p *MusicPlaylist) AddMusicPlayer(player player.MusicPlayerInterface) {
	p.players = append(p.players, player)
}

func (p *MusicPlaylist) GetItems() (items []ItemInterface) {
	return p.items
}

func (p *MusicPlaylist) GetCurrentItem() (item ItemInterface, remaining time.Duration) {
	item = p.currentItem
	switch p.status {
	case PLAYING:
		remaining = p.endTime.Sub(time.Now())
	case PAUSED:
		remaining = p.remainingDuration
	case STOPPED:
		if p.currentItem != nil {
			remaining = p.currentItem.GetDuration()
		}
	}
	return
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
		tempItem := plItem
		addedItems = append(addedItems, &tempItem)
	}
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	p.items = append(p.items, addedItems...)

	p.EmitEvent("items_added", addedItems)
	p.EmitEvent("list_updated", p.items)
	return
}

func (p *MusicPlaylist) ShuffleList() {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	for i := range p.items {
		j := rand.Intn(i + 1)
		p.items[i], p.items[j] = p.items[j], p.items[i]
	}
	p.EmitEvent("list_updated", p.items)
}

func (p *MusicPlaylist) EmptyList() {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	p.items = make([]ItemInterface, 0)
	p.EmitEvent("list_updated", p.items)
}

func (p *MusicPlaylist) GetStatus() (status Status) {
	return p.status
}

func (p *MusicPlaylist) Play() (item ItemInterface, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	switch p.status {
	case STOPPED:
		item, err = p.next()
		return
	case PAUSED:
		err = p.pause()
	}
	item = p.currentItem
	return
}

func (p *MusicPlaylist) playWait() {
	p.playTimer = time.NewTimer(p.endTime.Sub(time.Now()))

	// Wait for the timer to time out, or be canceled because of a STOP or something
	<-p.playTimer.C

	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	p.EmitEvent("play_done", p.currentItem)
	p.currentItem = nil

	if len(p.items) == 0 {
		p.stop()
	} else if len(p.items) > 0 && p.status == PLAYING {
		p.next()
	}
}

func (p *MusicPlaylist) Next() (item ItemInterface, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	return p.next()
}

func (p *MusicPlaylist) next() (item ItemInterface, err error) {
	if len(p.items) == 0 {
		err = errors.New("Playlist is empty, no next available")
		return
	}
	if p.status == PLAYING || p.status == PAUSED {
		p.stop()
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
	p.EmitEvent("play_start", p.currentItem)
	return
}

func (p *MusicPlaylist) Stop() (err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	return p.stop()
}

func (p *MusicPlaylist) stop() (err error) {
	if p.status == STOPPED || p.currentPlayer == nil {
		err = errors.New("Nothing currently playing")
		return
	}
	err = p.currentPlayer.Stop()
	if err != nil {
		err = fmt.Errorf("[%s] Error stopping: %v", p.currentPlayer.Name(), err)
		return
	}
	p.status = STOPPED
	p.currentItem = nil
	p.currentPlayer = nil
	if p.playTimer != nil {
		// Kill the current playWait()
		p.playTimer.Stop()
	}
	p.EmitEvent("stop")
	return
}

func (p *MusicPlaylist) Pause() (err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	return p.pause()
}

func (p *MusicPlaylist) pause() (err error) {
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

		p.EmitEvent("unpause", p.currentItem, p.remainingDuration)
	} else {
		p.status = PAUSED
		p.remainingDuration = p.endTime.Sub(time.Now())
		if p.playTimer != nil {
			// Kill the current playWait()
			p.playTimer.Stop()
		}
		p.EmitEvent("pause", p.currentItem, p.remainingDuration)
	}
	return
}
