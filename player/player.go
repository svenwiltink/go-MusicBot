package player

import (
	"errors"
	"fmt"
	"github.com/SvenWiltink/go-MusicBot/songplayer"
	"github.com/sirupsen/logrus"
	"github.com/vansante/go-event-emitter"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type Player struct {
	*eventemitter.Emitter

	queueStorage *QueueStorage

	currentSong      songplayer.Playable
	playlistPosition int
	playlist         []songplayer.Playable
	status           Status

	players       []songplayer.SongPlayer
	currentPlayer songplayer.SongPlayer

	playTimer         *time.Timer
	endTime           time.Time
	remainingDuration time.Duration

	controlMutex sync.Mutex
}

var ErrNothingPlaying = errors.New("nothing currently playing")

func NewPlayer(queueFilePath string) (player *Player) {
	player = &Player{
		Emitter:          eventemitter.NewEmitter(),
		queueStorage:     NewQueueStorage(queueFilePath),
		status:           STOPPED,
		playlistPosition: 0,
	}
	return
}

func (p *Player) Init() {
	urls, err := p.queueStorage.ReadQueue()
	if err != nil {
		p.EmitEvent("queue_error_loading", p.queueStorage.path, err)
		logrus.Warnf("Player.Init: Error reading queue from file [%s] %v", p.queueStorage.path, err)
	} else {
		var songs []songplayer.Playable
		for i, url := range urls {
			insertedSongs, err := p.insertSongs(url, i)
			if err != nil {
				logrus.Errorf("Player.Init: Error inserting song from queue [%d | %s] %v", i, url, err)
			}
			songs = append(songs, insertedSongs...)
		}

		logrus.Infof("Player.Init: Loaded %d songs from queue storage", len(songs))
		p.EmitEvent("queue_loaded", songs)
	}

	// Setup listener to keep the queue file updated
	p.AddListener("queue_updated", p.queueStorage.OnListUpdate)
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
	logrus.Infof("Player.AddSongPlayer: Songplayer %s added", player.Name())
	p.players = append(p.players, player)
}

func (p *Player) GetPastSongs() (songs []songplayer.Playable) {
	return p.playlist[:p.playlistPosition]
}

func (p *Player) GetQueuedSongs() (songs []songplayer.Playable) {
	if p.playlistPosition == len(p.playlist)-1 {
		return []songplayer.Playable{}
	}
	return p.playlist[p.playlistPosition+1:]
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

	addedSongs, err = p.insertSongs(url, len(p.GetQueuedSongs()))
	if err != nil {
		logrus.Warnf("Player.AddSongs: Error adding songs [%s] %v", url, err)
		return
	}
	logrus.Infof("Player.AddSongs: Added %d songs from url [%s]", len(addedSongs), url)
	return
}

func (p *Player) InsertSongs(url string, position int) (addedSongs []songplayer.Playable, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	addedSongs, err = p.insertSongs(url, position)
	if err != nil {
		logrus.Warnf("Player.InsertSongs: Error inserting songs [%s] %v", url, err)
		return
	}
	logrus.Infof("Player.InsertSongs: Inserted %d songs from url [%s]", len(addedSongs), url)
	return
}

func (p *Player) insertSongs(url string, position int) (addedSongs []songplayer.Playable, err error) {
	musicPlayer, err := p.findPlayer(url)
	if err != nil {
		logrus.Infof("Player.insertSongs: No songplayer found to play %s", url)
		return
	}

	addedSongs, err = musicPlayer.GetSongs(url)
	if err != nil {
		logrus.Warnf("Player.insertSongs: Error getting songs from url [%s] %v", musicPlayer.Name(), err)
		return
	}

	// Convert position by offsetting it against current position
	position += p.playlistPosition

	err = p.insertPlayables(addedSongs, position)
	return
}

func (p *Player) insertPlayables(playables []songplayer.Playable, position int) (err error) {
	if position < 0 || position > len(p.playlist) {
		err = errors.New("invalid position to insert songs")
		return
	}

	for i, playable := range playables {
		p.playlist = append(p.playlist, nil)
		copy(p.playlist[position+i+1:], p.playlist[position+i:])
		p.playlist[position+i] = playable
	}

	p.EmitEvent("songs_added", playables, position)
	p.EmitEvent("queue_updated", p.GetQueuedSongs())
	return
}

func (p *Player) ShuffleQueue() {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	for i := p.playlistPosition + 1; i < len(p.playlist); i++ {
		j := rand.Intn(i + 1)
		p.playlist[i], p.playlist[j] = p.playlist[j], p.playlist[i]
	}
	p.EmitEvent("queue_updated", p.GetQueuedSongs())

	logrus.Infof("Player.ShuffleQueue: Queue successfully shuffled")
}

func (p *Player) EmptyQueue() {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	newList := make([]songplayer.Playable, 0)
	// Copy over the play history
	for i := 0; i <= p.playlistPosition; i++ {
		newList = append(newList, p.playlist[i])
	}
	p.playlist = newList

	p.EmitEvent("queue_updated", p.GetQueuedSongs())

	logrus.Infof("Player.ShuffleQueue: Queue successfully emptied")
}

func (p *Player) GetStatus() (status Status) {
	return p.status
}

func (p *Player) Play() (song songplayer.Playable, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	switch p.status {
	case PAUSED:
		err = p.pause()
	default:
		song, err = p.setPlaylistPosition(p.playlistPosition)
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

	if p.playlistPosition < len(p.playlist)-1 && p.status == PLAYING {
		p.setPlaylistPosition(p.playlistPosition + 1)
	} else {
		p.stop()
	}
}

func (p *Player) Seek(positionSeconds int) (err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	if p.status == STOPPED || p.currentPlayer == nil {
		err = ErrNothingPlaying
		return
	}

	totalSeconds := int(p.currentSong.GetDuration().Seconds())
	if positionSeconds < 0 || positionSeconds > totalSeconds {
		err = fmt.Errorf("Position %d is out of bounds [0 - %d]", positionSeconds, totalSeconds)
		return
	}

	err = p.currentPlayer.Seek(positionSeconds)
	if err != nil {
		logrus.Warnf("Player.Seek: Error seeking to %d with player %s: %v", positionSeconds, p.currentPlayer.Name(), err)
		return
	}

	positionDuration := time.Duration(int64(time.Second) * int64(positionSeconds))
	remainingDuration := p.currentSong.GetDuration() - positionDuration

	p.playTimer.Reset(remainingDuration)
	p.endTime = time.Now().Add(remainingDuration)
	p.EmitEvent("song_seek", p.currentSong, remainingDuration)

	logrus.Infof("Player.Seek: Play resumed for remaining duration of %v", remainingDuration)
	return
}

func (p *Player) Next() (song songplayer.Playable, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	if p.playlistPosition+1 == len(p.playlist) {
		err = errors.New("no next available, queue is empty")
		return
	}

	return p.setPlaylistPosition(p.playlistPosition + 1)
}

func (p *Player) Previous() (song songplayer.Playable, err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	if p.playlistPosition == 0 {
		err = errors.New("no previous available, history is empty")
		return
	}

	return p.setPlaylistPosition(p.playlistPosition - 1)
}

func (p *Player) setPlaylistPosition(newPosition int) (song songplayer.Playable, err error) {
	if newPosition < 0 || len(p.playlist) == 0 || newPosition >= len(p.playlist) {
		err = errors.New("invalid playlist position")
		return
	}
	if p.status == PLAYING || p.status == PAUSED {
		p.stop()
	}

	song = p.playlist[newPosition]
	musicPlayer, err := p.findPlayer(song.GetURL())
	if err != nil {
		logrus.Errorf("Player.setPlaylistPosition: No player available to play [%s] %v", song.GetURL(), err)
		return
	}
	err = musicPlayer.Play(song.GetURL())
	if err != nil {
		logrus.Errorf("Player.setPlaylistPosition: Error playing %s with player %s: %v", song.GetURL(), musicPlayer.Name(), err)
		return
	}
	p.playlistPosition = newPosition
	p.currentSong = song
	p.currentPlayer = musicPlayer
	p.status = PLAYING
	p.endTime = time.Now().Add(song.GetDuration())

	// Start waiting for the song to be done
	go p.playWait()
	p.EmitEvent("play_start", p.currentSong)
	p.EmitEvent("queue_updated", p.GetQueuedSongs())

	logrus.Infof("Player.setPlaylistPosition: %s started playing %s successfully", musicPlayer.Name(), song.GetURL())
	return
}

func (p *Player) Stop() (err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	return p.stop()
}

func (p *Player) stop() (err error) {
	if p.status == STOPPED || p.currentPlayer == nil {
		err = ErrNothingPlaying
		return
	}
	currentPlayer := p.currentPlayer
	err = currentPlayer.Stop()
	if err != nil {
		logrus.Errorf("Player.stop: Error stopping player %s: %v", currentPlayer.Name(), err)
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

	logrus.Infof("Player.stop: %s stopped playing", currentPlayer.Name())
	return
}

func (p *Player) Pause() (err error) {
	p.controlMutex.Lock()
	defer p.controlMutex.Unlock()

	return p.pause()
}

func (p *Player) pause() (err error) {
	if p.status == STOPPED || p.currentPlayer == nil {
		err = ErrNothingPlaying
		return
	}

	err = p.currentPlayer.Pause(p.status != PAUSED)
	if err != nil {
		logrus.Errorf("Player.pause: Error (un)pausing player %s [%v]: %v", p.currentPlayer.Name(), p.status != PAUSED, err)
		return
	}
	if p.status == PAUSED {
		p.status = PLAYING
		p.endTime = time.Now().Add(p.remainingDuration)
		// Restart the play wait goroutine with the new time
		go p.playWait()

		p.EmitEvent("unpause", p.currentSong, p.remainingDuration)

		logrus.Infof("Player.pause: %s resumed playing %s", p.currentPlayer.Name(), p.currentSong.GetURL())
	} else {
		p.status = PAUSED
		p.remainingDuration = p.endTime.Sub(time.Now())
		if p.playTimer != nil {
			// Kill the current playWait()
			p.playTimer.Stop()
		}
		p.EmitEvent("pause", p.currentSong, p.remainingDuration)

		logrus.Infof("Player.pause: %s paused playing %s", p.currentPlayer.Name(), p.currentSong.GetURL())
	}
	return
}
