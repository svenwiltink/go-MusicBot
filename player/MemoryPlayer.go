package player

import (
	"time"
	"fmt"
)

type PlayTimer struct {
	*time.Timer

	end time.Time
}

type MemoryPlayer struct {
	Queue        Queue
	CurrentSong  QueueItem
	Status       Status

	timer     *PlayTimer
	remaining time.Duration
}

func NewMemoryPlayer() MusicPlayer {
	player := &MemoryPlayer{
		Queue:        NewQueue(),
		Status:       STOPPED,
	}

	player.init()

	return player
}

func (p *MemoryPlayer) init() {
	fmt.Print("MemoryPlayer - Init\n")

}

func (p *MemoryPlayer) GetStatus() Status {
	return p.Status
}

func (p *MemoryPlayer) Start() error {
	fmt.Print("MemoryPlayer - Start\n")

	if p.Status == STOPPED {
		p.Status = RUNNING

		p.Next()
	}

	return nil
}

func (p *MemoryPlayer) Stop() {
	fmt.Print("MemoryPlayer - Stop\n")

	if p.Status == RUNNING {
		p.Status = STOPPED
		p.timer.Stop()
	}
}

func (p *MemoryPlayer) Pause() {
	fmt.Print("MemoryPlayer - Pause\n")
	if p.Status == STOPPED {
		return
	}

	p.remaining = p.timer.end.Sub(time.Now())
	p.timer.Stop()
}

func (p *MemoryPlayer) Next() {
	fmt.Print("MemoryPlayer - Next\n")
	if p.timer != nil {
		p.timer.Stop()
	}

	item, _ := p.Queue.shift()
	p.remaining = item.Duration

	p.Play()

}

func (p *MemoryPlayer) Play() {
	fmt.Print("MemoryPlayer - Play\n")

	if p.Status == RUNNING {
		return
	}

	// play the song :D
	p.timer = &PlayTimer{ time.NewTimer(p.remaining), time.Now().Add(p.remaining)}
	go func() {
		<- p.timer.C
		p.Next()
	}()
}

func (p *MemoryPlayer) AddSong(source string, duration int64) {

	p.Queue.add(QueueItem{
		URL: source,
		Duration: time.Duration(duration) * time.Second,
	})

	if p.Status == STOPPED {
		p.Start()
	}
}

func (p *MemoryPlayer) GetCurrentSong() *QueueItem {
	return &QueueItem{}
}

func (p *MemoryPlayer) GetQueueItems() []QueueItem {
	return []QueueItem{{}}
}

func (p *MemoryPlayer) FlushQueue() {

}

func (p *MemoryPlayer) ShuffleQueue() {

}