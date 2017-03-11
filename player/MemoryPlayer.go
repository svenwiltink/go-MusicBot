package player

import (
	"fmt"
	"time"
)

type PlayTimer struct {
	*time.Timer

	end time.Time
}

type MemoryPlayer struct {
	Queue       Queue
	CurrentSong *QueueItem
	Status      Status

	timer     *PlayTimer
	remaining time.Duration
}

func NewMemoryPlayer() (p MusicPlayer) {
	p = &MemoryPlayer{
		Queue:  NewQueue(),
		Status: STOPPED,
	}

	p.Init()

	return
}

// Init - Initialize the player
func (p *MemoryPlayer) Init() {
	fmt.Print("MemoryPlayer - Init\n")

}

// GetStatus - Get the current status of the player
func (p *MemoryPlayer) GetStatus() Status {
	return p.Status
}

// Start - Start the player
func (p *MemoryPlayer) Start() error {
	fmt.Print("MemoryPlayer - Start\n")

	if p.Status == STOPPED {
		p.Status = RUNNING

		p.Next()
	}

	return nil
}

// Stop - Stop playing
func (p *MemoryPlayer) Stop() {
	fmt.Print("MemoryPlayer - Stop\n")

	if p.Status == RUNNING {
		p.Status = STOPPED
		p.timer.Stop()
	}
}

// Pause - Pause playing
func (p *MemoryPlayer) Pause() {
	fmt.Print("MemoryPlayer - Pause\n")

	if p.Status == STOPPED {
		p.Play()
		return
	}

	p.remaining = p.timer.end.Sub(time.Now())
	p.Stop()
}

// Next - Get and play the next item in the queue
func (p *MemoryPlayer) Next() {
	fmt.Print("MemoryPlayer - Next\n")

	if p.timer != nil {
		p.timer.Stop()
	}

	item, err := p.Queue.shift()
	if err != nil {
		p.Stop()
		return
	}

	p.remaining = item.Duration

	p.Play()

}

// Play -
func (p *MemoryPlayer) Play() {
	fmt.Print("MemoryPlayer - Play\n")

	// play the song :D
	p.timer = &PlayTimer{time.NewTimer(p.remaining), time.Now().Add(p.remaining)}
	go func() {
		<-p.timer.C
		p.Next()
	}()
}

// AddSong - Add a song to the queue
func (p *MemoryPlayer) AddSong(source string, duration int64) {
	fmt.Printf("MemoryPlayer - AddSong - %s %d\n", source, duration)

	p.Queue.add(QueueItem{
		URL:      source,
		Duration: time.Duration(duration) * time.Second,
	})

	if p.Status == STOPPED {
		p.Start()
	}
}

func (p *MemoryPlayer) GetCurrentSong() *QueueItem {
	return p.CurrentSong
}

func (p *MemoryPlayer) GetQueueItems() []QueueItem {
	return p.Queue.Items
}

func (p *MemoryPlayer) FlushQueue() {
	p.Queue.Flush()
}

func (p *MemoryPlayer) ShuffleQueue() {
	p.Queue.Shuffle()
}
