package player

type Status int

const (
	RUNNING Status = 1 + iota
	STOPPED
)

type MusicPlayer interface {
	GetStatus() Status
	Start() error
	Stop()
	Pause()
	Next()
	AddSong(string, int64)
	GetCurrentSong() *QueueItem
	GetQueueItems() []QueueItem
	FlushQueue()
	ShuffleQueue()
}
