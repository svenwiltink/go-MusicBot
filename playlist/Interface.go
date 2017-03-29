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
	AddItems(url string) (items []ItemInterface, err error)
	ShuffleList()
	EmptyList()
	GetStatus() (status Status)
	Play() (err error)
	Next() (item ItemInterface, err error)
	Stop() (err error)
	Pause() (err error)
}

type ItemInterface interface {
	Title() string
	Duration() time.Duration
	URL() string
}
