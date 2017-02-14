package player

import (
	"errors"
	"math/rand"
)

type Queue struct {
	Items []QueueItem
}

func (q *Queue) HasNext() bool {
	return len(q.Items) > 0
}

func (q *Queue) shift() (QueueItem, error) {
	if len(q.Items) == 0 {
		return QueueItem{}, errors.New("No next song available")
	}

	item, remainder := q.Items[0], q.Items[1:]
	q.Items = remainder
	return item, nil
}

func (q *Queue) add(item QueueItem) {
	q.Items = append(q.Items, item)
}

// Shuffle - Shuffle all items in the queue
func (q *Queue) Shuffle() {
	for i := range q.Items {
		j := rand.Intn(i + 1)
		q.Items[i], q.Items[j] = q.Items[j], q.Items[i]
	}
}

// Flush - Flush the entire queue
func (q *Queue) Flush() {
	q.Items = make([]QueueItem, 0)
}

func NewQueue() Queue {
	return Queue{
		Items: make([]QueueItem, 0),
	}
}

type QueueItem struct {
	URL string
}

func NewQueueItem(URL string) QueueItem {
	return QueueItem{
		URL: URL,
	}
}
func (i *QueueItem) GetURL() string {
	return i.URL
}
