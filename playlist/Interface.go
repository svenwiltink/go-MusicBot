package playlist

import "time"

type Status int

const (
	PLAYING Status = 1 + iota
	PAUSED
	STOPPED
)

type ListInterface interface {
	GetItems() (items []ItemInterface)
	GetCurrentItem() (item ItemInterface)
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
