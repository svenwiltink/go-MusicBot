package player

import (
	"time"
)

type MemoryPlayer struct{}

func NewMemoryPlayer() (p *MemoryPlayer) {
	p = &MemoryPlayer{}
	return
}

func (p *MemoryPlayer) Name() (name string) {
	return "MemoryPlayer"
}

func (p *MemoryPlayer) CanPlay(url string) (canPlay bool) {
	// Sure, we can play anything you want
	return true
}

func (p *MemoryPlayer) GetItems(url string) (items []ListItem, err error) {
	items = append(items, *NewListItem(url, time.Minute, url))
	return
}

func (p *MemoryPlayer) Play(url string) (err error) {
	// Do nothing \o/
	return
}

func (p *MemoryPlayer) Pause(pauseState bool) (err error) {
	// Do nothing \o/
	return
}

func (p *MemoryPlayer) Stop() (err error) {
	// Do nothing \o/
	return
}
