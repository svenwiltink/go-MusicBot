package player

import (
	"fmt"
	"github.com/vansante/go-spotify-control"
	"time"
)

type SpotifyPlayer struct {
	Queue       Queue
	CurrentSong *QueueItem
	Control     *spotifycontrol.SpotifyControl

	remaining time.Duration
}

func NewSpotifyPlayer() (p MusicPlayer) {
	p = &SpotifyPlayer{
		Queue: NewQueue(),
	}

	p.Init()

	return
}

// Init - Initialize the player
func (p *SpotifyPlayer) Init() {
	fmt.Println("SpotifyPlayer - Init")

	var err error
	p.Control, err = spotifycontrol.NewSpotifyControl("", 0)
	if err != nil {
		fmt.Println(err)
	}
}

// GetStatus - Get the current status of the player
func (p *SpotifyPlayer) GetStatus() Status {
	status, err := p.Control.GetStatus()
	if err != nil {
		fmt.Println(err)
		return STOPPED
	}
	if !status.Playing {
		return STOPPED
	}
	return RUNNING
}

// Start - Start the player
func (p *SpotifyPlayer) Start() error {
	fmt.Println("SpotifyPlayer - Start")

	if p.GetStatus() == STOPPED {
		p.Next()
	}

	return nil
}

// Stop - Stop playing
func (p *SpotifyPlayer) Stop() {
	fmt.Println("SpotifyPlayer - Stop")
	p.Control.Pause()
}

// Pause - Pause playing
func (p *SpotifyPlayer) Pause() {
	fmt.Println("SpotifyPlayer - Pause")

	if p.GetStatus() == STOPPED {
		p.Control.Unpause()
	} else {
		p.Control.Pause()
	}
}

// Next - Get and play the next item in the queue
func (p *SpotifyPlayer) Next() {
	fmt.Println("SpotifyPlayer - Next")

	if !p.Queue.HasNext() {
		p.CurrentSong = nil
		return
	}

	nextSong, err := p.Queue.shift()
	if err != nil {
		fmt.Println(err)
		return
	}

	p.CurrentSong = &nextSong
	url := nextSong.GetURL()
	status, err := p.Control.Play(url)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		time.Sleep(time.Duration(status.Track.Length) * time.Millisecond)

		p.Next()
	}()
}

// AddSong - Add a song to the queue
func (p *SpotifyPlayer) AddSong(source string) {
	fmt.Printf("SpotifyPlayer - AddSong - %s 60\n", source)

	p.Queue.add(QueueItem{
		URL: source,
	})

	if p.GetStatus() == STOPPED {
		p.Start()
	}
}

func (p *SpotifyPlayer) GetCurrentSong() *QueueItem {
	return p.CurrentSong
}

func (p *SpotifyPlayer) GetQueueItems() []QueueItem {
	return p.Queue.Items
}

func (p *SpotifyPlayer) FlushQueue() {
	p.Queue.Flush()
}

func (p *SpotifyPlayer) ShuffleQueue() {
	p.Queue.Shuffle()
}
