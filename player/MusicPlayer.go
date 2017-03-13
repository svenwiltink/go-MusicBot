package player

type Status int

const (
	RUNNING Status = 1 + iota
	STOPPED
)

type MusicPlayer interface {
	GetStatus() Status
	Init()
	Start() error
	Stop()
	Pause()
	Next()
	AddSong(string)
	GetCurrentSong() *QueueItem
	GetQueueItems() []QueueItem
	FlushQueue()
	ShuffleQueue()
}
