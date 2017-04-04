package player

import (
	"errors"
	"fmt"
	"github.com/vansante/go-event-emitter"
	"gitlab.transip.us/swiltink/go-MusicBot/songplayer"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Player struct {
	*eventemitter.Emitter

	currentSong songplayer.Playable
	queue       []songplayer.Playable
	status      Status

	players       []songplayer.SongPlayer
	currentPlayer songplayer.SongPlayer

	playTimer         *time.Timer
	endTime           time.Time
	remainingDuration time.Duration

	controlMutex sync.Mutex
}

func NewPlayer() (player *Player) {
	player = &Player{
		Emitter: eventemitter.NewEmitter(),
		status:  STOPPED,
	}
	return
}

func (p *Player) GetSongPlayer(name string) (songPlayer songplayer.SongPlayer) {
	for _, plr := range p.players {
		if strings.ToLower(plr.Name()) == strings.ToLower(name) {
			songPlayer = plr
			return
		}
	}
	return
}

func (p *Player) GetSongPlayers() (players []songplayer.SongPlayer) {
	return p.players
}

func (p *Player) AddSongPlayer(player songplayer.SongPlayer) {
	p.players = append(p.players, player)
}

func (p *Player) GetQueuedSongs() (songs []songplayer.Playable) {
	return p.queue
}

func (p *Player) GetCurrentSong() (song songplayer.Playable, remaining time.Duration) {
	song = p.currentSong
	switch p.status {
	case PLAYING:
		remaining = p.endTime.Sub(time.Now())
	case PAUSED:
		remaining = p.remainingDuration
	case STOPPED:
		if p.currentSong != nil {
			remaining = p.currentSong.GetDuration()
		}
	}
	return
}

func (p *Player) findPlayer(url string) (songPlayer songplayer.SongPlayer, err error) {
	for _, play := range p.players {
		if play.CanPlay(url) {
			songPlayer = play
			return
		}
	}
	err = fmt.Errorf("No suitable songplayer found to play %s", url)
	return
}

func (p *Player) AddSongs(url string) (addedSongs []songplayer.Playable, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	return p.insertSongs(url, len(p.queue))
}

func (p *Player) InsertSongs(url string, position int) (addedSongs []songplayer.Playable, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	if position < 0 || position > len(p.queue) {
		err = errors.New("invalid position to insert items")
		return
	}

	return p.insertSongs(url, position)
}

func (p *Player) insertSongs(url string, position int) (addedSongs []songplayer.Playable, err error) {
	musicPlayer, err := p.findPlayer(url)
	if err != nil {
		return
	}

	addedSongs, err = musicPlayer.GetSongs(url)
	if err != nil {
		err = fmt.Errorf("[%s] Error getting items from url: %v", musicPlayer.Name(), err)
		return
	}

	for i, addSong := range addedSongs {
		p.queue = append(p.queue, nil)
		copy(p.queue[position+i+1:], p.queue[position+i:])
		p.queue[position+i] = addSong
	}

	p.EmitEvent("items_added", addedSongs)
	p.EmitEvent("list_updated", p.queue)
	return
}

func (p *Player) ShuffleQueue() {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	for i := range p.queue {
		j := rand.Intn(i + 1)
		p.queue[i], p.queue[j] = p.queue[j], p.queue[i]
	}
	p.EmitEvent("list_updated", p.queue)
}

func (p *Player) EmptyQueue() {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	p.queue = make([]songplayer.Playable, 0)
	p.EmitEvent("list_updated", p.queue)
}

func (p *Player) GetStatus() (status Status) {
	return p.status
}

func (p *Player) Play() (song songplayer.Playable, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	switch p.status {
	case STOPPED:
		song, err = p.next()
		return
	case PAUSED:
		err = p.pause()
	}
	song = p.currentSong
	return
}

func (p *Player) playWait() {
	p.playTimer = time.NewTimer(p.endTime.Sub(time.Now()))

	// Wait for the timer to time out, or be canceled because of a STOP or something
	<-p.playTimer.C

	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	p.EmitEvent("play_done", p.currentSong)
	p.currentSong = nil

	if len(p.queue) == 0 {
		p.stop()
	} else if len(p.queue) > 0 && p.status == PLAYING {
		p.next()
	}
}

func (p *Player) Next() (song songplayer.Playable, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	return p.next()
}

func (p *Player) next() (song songplayer.Playable, err error) {
	if len(p.queue) == 0 {
		err = errors.New("Playlist is empty, no next available")
		return
	}
	if p.status == PLAYING || p.status == PAUSED {
		p.stop()
	}

	song, p.queue = p.queue[0], p.queue[1:]
	musicPlayer, err := p.findPlayer(song.GetURL())
	if err != nil {
		return
	}
	err = musicPlayer.Play(song.GetURL())
	if err != nil {
		err = fmt.Errorf("[%s] Error playing: %v", musicPlayer.Name(), err)
		return
	}
	p.currentSong = song
	p.currentPlayer = musicPlayer
	p.status = PLAYING
	p.endTime = time.Now().Add(song.GetDuration())
	// Start waiting for the song to be done
	go p.playWait()
	p.EmitEvent("play_start", p.currentSong)
	return
}

func (p *Player) Stop() (err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	return p.stop()
}

func (p *Player) stop() (err error) {
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
	p.currentSong = nil
	p.currentPlayer = nil
	if p.playTimer != nil {
		// Kill the current playWait()
		p.playTimer.Stop()
	}
	p.EmitEvent("stop")
	return
}

func (p *Player) Pause() (err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	return p.pause()
}

func (p *Player) pause() (err error) {
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

		p.EmitEvent("unpause", p.currentSong, p.remainingDuration)
	} else {
		p.status = PAUSED
		p.remainingDuration = p.endTime.Sub(time.Now())
		if p.playTimer != nil {
			// Kill the current playWait()
			p.playTimer.Stop()
		}
		p.EmitEvent("pause", p.currentSong, p.remainingDuration)
	}
	return
}
