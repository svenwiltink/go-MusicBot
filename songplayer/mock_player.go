package songplayer

import (
	"time"
)

type MockPlayer struct{}

func NewMemoryPlayer() (p *MockPlayer) {
	p = &MockPlayer{}
	return
}

func (p *MockPlayer) Name() (name string) {
	return "MockPlayer"
}

func (p *MockPlayer) CanPlay(url string) (canPlay bool) {
	// Sure, we can play anything you want
	return true
}

func (p *MockPlayer) GetItems(url string) (items []Song, err error) {
	items = append(items, *NewSong(url, time.Minute, url, ""))
	return
}

func (p *MockPlayer) Play(url string) (err error) {
	// Do nothing \o/
	return
}

func (p *MockPlayer) Seek(positionSeconds int) (err error) {
	// Do nothing \o/
	return
}

func (p *MockPlayer) Pause(pauseState bool) (err error) {
	// Do nothing \o/
	return
}

func (p *MockPlayer) Stop() (err error) {
	// Do nothing \o/
	return
}
