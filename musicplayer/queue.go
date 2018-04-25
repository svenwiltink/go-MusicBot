package musicplayer

import (
	"log"
	"sync"

	"github.com/svenwiltink/go-musicbot/musicplayer/musicprovider"
)

// Queue holds an array of songs
type Queue struct {
	songs         []*musicprovider.Song
	lock          sync.Mutex
	songAddedChan chan bool
}

func (queue *Queue) append(songs ...*musicprovider.Song) {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	queue.songs = append(queue.songs, songs...)
	log.Println("Song added to the queue")
	queue.notifyWaiting()
}

func (queue *Queue) notifyWaiting() {
	select {
	case queue.songAddedChan <- true:
		log.Println("Notified a waiting queue listener")
		break
	default:
		break
	}
}

// GetNext returns the next item in the queue if it exists
func (queue *Queue) GetNext() *musicprovider.Song {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	if len(queue.songs) == 0 {
		return nil
	}
	song, remaining := queue.songs[0], queue.songs[1:]

	queue.songs = remaining

	return song
}

// WaitForNext is a blocking call that returns the next song in the queue and wait for one to be added
// if there is no song available. The only caveat of this is that this method can only really be used by one process
// because it uses a single signal
func (queue *Queue) WaitForNext() *musicprovider.Song {
	next := queue.GetNext()

	if next != nil {
		return next
	}

	log.Println("Waiting for a song to be added")
	<-queue.songAddedChan
	return queue.GetNext()
}

// NewQueue creates a new instance of Queue
func NewQueue() *Queue {
	return &Queue{
		songs:         make([]*musicprovider.Song, 0),
		songAddedChan: make(chan bool),
	}
}
