package music

import (
	"log"
	"sync"
	
	"github.com/vansante/go-event-emitter"
	"time"
	"errors"
)

const (
	songAdded eventemitter.EventType = "song-added"
)

// Queue holds an array of songs
type Queue struct {
	*eventemitter.Emitter
	songs         []*Song
	lock          sync.Mutex
}

func (queue *Queue) Append(songs ...*Song) {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	queue.songs = append(queue.songs, songs...)
	log.Println("Song added to the queue")
	queue.EmitEvent(songAdded)
}

// GetNext returns the next item in the queue if it exists
func (queue *Queue) GetNext() *Song {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	return queue.getNext()
}

func (queue *Queue) getNext() *Song {
	if len(queue.songs) == 0 {
		return nil
	}
	song, remaining := queue.songs[0], queue.songs[1:]

	queue.songs = remaining

	return song
}

func (queue *Queue) GetLength() int{
	queue.lock.Lock()
	defer queue.lock.Unlock()

	return len(queue.songs)
}

func (queue *Queue) GetTotalDuration() time.Duration {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	var duration time.Duration

	for _, song := range queue.songs {
		duration = duration + song.Duration
	}

	return duration.Round(time.Second)
}

func (queue *Queue) GetNextN(limit int) ([]Song, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}

	queue.lock.Lock()
	defer queue.lock.Unlock()

	if len(queue.songs) < limit {
		limit = len(queue.songs)
	}

	result := make([]Song, limit)
	for i := 0; i < limit; i++ {
		result[i] = *queue.songs[i]
	}

	return result, nil
}

// WaitForNext is a blocking call that returns the next song in the queue and wait for one to be added
// if there is no song available.
func (queue *Queue) WaitForNext() *Song {
	next := queue.GetNext()

	if next != nil {
		return next
	}

	done := make(chan struct{})
	queue.ListenOnce(songAdded, func(args ...interface{}) {
		done <- struct{}{}
	})

	<-done
	return queue.GetNext()
}

// NewQueue creates a new instance of Queue
func NewQueue() *Queue {
	return &Queue{
		songs:         make([]*Song, 0),
		Emitter:       eventemitter.NewEmitter(true),
	}
}
