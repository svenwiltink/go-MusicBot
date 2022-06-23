package music

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestQueue_GetLength(t *testing.T) {
	t.Parallel()
	queue := NewQueue()

	assert.Equal(t, 0, queue.GetLength())

	queue.Append(Song{
		Duration: time.Minute,
		Name:     "banaan",
	})

	assert.Equal(t, 1, queue.GetLength())
}

func TestQueue_Append(t *testing.T) {
	t.Parallel()
	queue := NewQueue()

	song := Song{
		Duration: time.Minute,
		Name:     "banaan",
	}

	queue.Append(song)

	assert.Equal(t, queue.songs[0], song)
}

func TestQueue_Delete(t *testing.T) {
	newQueue := func() *Queue {
		queue := NewQueue()
		queue.Append(Song{Name: "banaan1"})
		queue.Append(Song{Name: "banaan2"})
		queue.Append(Song{Name: "banaan3"})

		return queue
	}

	t.Run("delete first item", func(t *testing.T) {
		t.Parallel()

		queue := newQueue()
		err := queue.Delete(0)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, 2, len(queue.songs))
		assert.Equal(t, "banaan2", queue.songs[0].Name)
		assert.Equal(t, "banaan3", queue.songs[1].Name)
	})

	t.Run("delete middle item", func(t *testing.T) {
		t.Parallel()

		queue := newQueue()
		err := queue.Delete(1)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, 2, len(queue.songs))
		assert.Equal(t, "banaan1", queue.songs[0].Name)
		assert.Equal(t, "banaan3", queue.songs[1].Name)
	})

	t.Run("delete last item", func(t *testing.T) {
		t.Parallel()

		queue := newQueue()
		err := queue.Delete(2)
		if !assert.NoError(t, err) {
			return
		}

		assert.Equal(t, 2, len(queue.songs))
		assert.Equal(t, "banaan1", queue.songs[0].Name)
		assert.Equal(t, "banaan2", queue.songs[1].Name)
	})

	t.Run("delete invalid item", func(t *testing.T) {
		t.Parallel()

		queue := newQueue()
		err := queue.Delete(6)
		if assert.Error(t, err) {
			assert.Equal(t, ErrQueueItemNotAvailable, err)
		}
	})

	t.Run("delete negative item", func(t *testing.T) {
		t.Parallel()

		queue := NewQueue()
		err := queue.Delete(-1)

		if assert.Error(t, err) {
			assert.Equal(t, "can not remove negative queue item -1", err.Error())
		}
	})
}

func TestQueue_Flush(t *testing.T) {
	t.Parallel()
	queue := NewQueue()

	song := Song{
		Duration: time.Minute,
		Name:     "banaan",
	}

	queue.Append(song)
	assert.NotEmpty(t, queue.songs)

	queue.Flush()
	assert.Empty(t, queue.songs)
}

func TestQueue_Append_Multiple(t *testing.T) {
	t.Parallel()
	queue := NewQueue()

	queue, song1, song2 := getTestQueue()

	assert.Equal(t, queue.songs[0], song1)
	assert.Equal(t, queue.songs[1], song2)
}

func TestQueue_GetNext(t *testing.T) {
	t.Parallel()
	queue := NewQueue()
	song := Song{
		Duration: time.Minute,
		Name:     "banaan",
	}

	queue.Append(song)
	queuedSong, _ := queue.getNext()
	assert.Equal(t, song, queuedSong)
}

func TestQueue_GetNext_Empty(t *testing.T) {
	t.Parallel()
	queue := NewQueue()

	_, err := queue.GetNext()
	if assert.Error(t, err) {
		assert.Equal(t, "no song available", err.Error())
	}
}

func TestQueue_GetNextN_Negative(t *testing.T) {
	t.Parallel()

	queue := NewQueue()

	_, err := queue.GetNextN(-1)
	assert.Error(t, err)
}

func TestQueue_GetNextN_Zero(t *testing.T) {
	t.Parallel()

	queue := NewQueue()

	_, err := queue.GetNextN(0)
	assert.Error(t, err)
}

func TestQueue_GetNextN_Empty(t *testing.T) {
	t.Parallel()

	queue := NewQueue()

	songs, err := queue.GetNextN(1)

	assert.NoError(t, err)
	assert.Empty(t, songs)
}

func TestQueue_GetNextN(t *testing.T) {
	t.Parallel()
	queue, song1, song2 := getTestQueue()

	nextN, err := queue.GetNextN(2)

	assert.NoError(t, err)
	assert.Equal(t, []Song{song1, song2}, nextN)
}

func TestQueue_GetTotalDuration(t *testing.T) {
	queue, _, _ := getTestQueue()

	assert.Equal(t, 2*time.Minute, queue.GetTotalDuration())
}

func TestQueue_Shuffle(t *testing.T) {
	queue, _, _ := getTestQueue()

	queue.Append(Song{
		Duration: time.Minute,
		Name:     "song3",
	})

	queue.randSource = rand.New(rand.NewSource(1))
	original := append([]Song(nil), queue.songs...)

	queue.Shuffle()

	assert.NotEqual(t, original, queue.songs)
}

func TestQueue_WaitForNext(t *testing.T) {
	t.Parallel()

	queue := NewQueue()

	song := Song{
		Duration: time.Minute,
		Name:     "song1",
	}

	go func() {
		queue.Append(song)
	}()

	assert.Equal(t, song, queue.WaitForNext())
}

func getTestQueue() (*Queue, Song, Song) {
	queue := NewQueue()
	song1 := Song{
		Duration: time.Minute,
		Name:     "song1",
	}
	song2 := Song{
		Duration: time.Minute,
		Name:     "song2",
	}
	queue.Append(song1, song2)
	return queue, song1, song2
}
