package MusicPlayer

import (
	"errors"
)

type Queue struct {
	Items []QueueItem
}

func (q *Queue) HasNext() bool{
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

func NewQueue() Queue {
	return Queue{
		Items: make([]QueueItem, 0),
	}
}

type QueueItem struct {
	Url string
}

func NewQueueItem(Url string) QueueItem {
	return QueueItem{
		Url: Url,
	}
}
func (i *QueueItem) GetUrl() string  {
	return i.Url
}