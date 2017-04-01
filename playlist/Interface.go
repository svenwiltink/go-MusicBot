package playlist

import (
	"github.com/vansante/go-event-emitter"
	"time"
)

type Status int

const (
	PLAYING Status = 1 + iota
	PAUSED
	STOPPED
)

type ListInterface interface {
	eventemitter.Observable

	GetItems() (items []ItemInterface)
	GetCurrentItem() (item ItemInterface, remaining time.Duration)
	AddItems(url string) (addedItems []ItemInterface, err error)
	ShuffleList()
	EmptyList()
	GetStatus() (status Status)
	Play() (item ItemInterface, err error)
	Next() (item ItemInterface, err error)
	Stop() (err error)
	Pause() (err error)
}

type ItemInterface interface {
	GetTitle() string
	GetDuration() time.Duration
	GetURL() string
}
