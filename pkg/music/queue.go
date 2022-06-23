package music

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"errors"
	"github.com/vansante/go-event-emitter"
	"time"
)

const (
	songAdded   eventemitter.EventType = "song-added"
	songDeleted eventemitter.EventType = "song-deleted"
)

var (
	ErrNoSongAvailable = errors.New("no song available")
)

// Queue holds an array of songs
type Queue struct {
	*eventemitter.Emitter
	songs      []Song
	lock       sync.Mutex
	randSource *rand.Rand
}

func (queue *Queue) Append(songs ...Song) {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	queue.songs = append(queue.songs, songs...)
	log.Println("Song added to the queue")
	queue.EmitEvent(songAdded)
}

func (queue *Queue) Delete(item int) error {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	if item < 1 {
		return fmt.Errorf("can not remove negative queue item %d", item)
	}

	if len(queue.songs)+1 < item {
		return fmt.Errorf("queue-item %d is not in queue", item)
	}

	queue.songs = append(queue.songs[:item-1], queue.songs[item:]...)
	log.Println("Song deleted from the queue")
	queue.EmitEvent(songDeleted)

	return nil
}

// GetNext returns the next item in the queue if it exists
func (queue *Queue) GetNext() (Song, error) {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	return queue.getNext()
}

func (queue *Queue) getNext() (Song, error) {
	if len(queue.songs) == 0 {
		return Song{}, ErrNoSongAvailable
	}
	song, remaining := queue.songs[0], queue.songs[1:]

	queue.songs = remaining

	return song, nil
}

func (queue *Queue) GetLength() int {
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
		result[i] = queue.songs[i]
	}

	return result, nil
}

func (queue *Queue) Shuffle() {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	// Shuffle numbers, swapping corresponding entries in letters at the same time.
	queue.randSource.Shuffle(len(queue.songs), func(i, j int) {
		queue.songs[i], queue.songs[j] = queue.songs[j], queue.songs[i]
	})
}

func (queue *Queue) Flush() {
	queue.lock.Lock()
	defer queue.lock.Unlock()
	queue.songs = make([]Song, 0)
}

// WaitForNext is a blocking call that returns the next song in the queue and wait for one to be added
// if there is no song available.
func (queue *Queue) WaitForNext() Song {
	next, err := queue.GetNext()

	if err == nil {
		return next
	}

	// keep waiting until we have a song. This makes sure we keep waiting if there was a race between 2 calls
	// to WaitForNext
	for {
		done := make(chan struct{})
		queue.ListenOnce(songAdded, func(args ...interface{}) {
			done <- struct{}{}
		})

		<-done

		song, err := queue.getNext()

		// we actually have a song now :D
		if err == nil {
			return song
		}
	}
}

// NewQueue creates a new instance of Queue
func NewQueue() *Queue {
	return &Queue{
		songs:      make([]Song, 0),
		Emitter:    eventemitter.NewEmitter(true),
		randSource: rand.New(rand.NewSource(time.Now().UTC().UnixNano())),
	}
}
