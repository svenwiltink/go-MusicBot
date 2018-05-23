package player

import (
	"log"
	"sync"

	"github.com/svenwiltink/go-musicbot/music"
	"github.com/vansante/go-event-emitter"
	"fmt"
)


const (
	songAdded eventemitter.EventType = "song-added"
)

// Queue holds an array of songs
type Queue struct {
	*eventemitter.Emitter
	songs         []*music.Song
	lock          sync.Mutex
	songAddedChan chan bool
}

func (queue *Queue) append(songs ...*music.Song) {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	queue.songs = append(queue.songs, songs...)
	log.Println("Song added to the queue")
	queue.EmitEvent(songAdded)
}

// GetNext returns the next item in the queue if it exists
func (queue *Queue) GetNext() *music.Song {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	return queue.getNext()
}

func (queue *Queue) getNext() *music.Song {
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
func (queue *Queue) WaitForNext() *music.Song {
	next := queue.GetNext()

	if next != nil {
		return next
	}

	done := make(chan struct{})
	queue.ListenOnce(songAdded, func(args ...interface{}) {
		done <- struct{}{}
	})

	<- done
	return queue.GetNext()
}

// NewQueue creates a new instance of Queue
func NewQueue() *Queue {
	return &Queue{
		songs:         make([]*music.Song, 0),
		songAddedChan: make(chan bool),
		Emitter: eventemitter.NewEmitter(true),
	}
}
