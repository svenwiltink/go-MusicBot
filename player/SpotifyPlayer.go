package player

import (
	"github.com/vansante/go-spotify-control"
	"strings"
	"time"
)

type SpotifyPlayer struct {
	control *spotifycontrol.SpotifyControl
}

func NewSpotifyPlayer() (p *SpotifyPlayer, err error) {
	cntrl, err := spotifycontrol.NewSpotifyControl("", 1*time.Second)
	if err != nil {
		return
	}

	p = &SpotifyPlayer{
		control: cntrl,
	}

	return
}

func (p *SpotifyPlayer) Name() (name string) {
	return "SpotifyPlayer"
}

func (p *SpotifyPlayer) CanPlay(url string) (canPlay bool) {
	return strings.Contains(strings.ToLower(url), "spotify")
}

func (p *SpotifyPlayer) GetItems(url string) (items []ListItem, err error) {
	// TODO: FIXME: Get meta data, add each song individually, etc
	items = append(items, *NewListItem(url, time.Minute, url))
	return
}

func (p *SpotifyPlayer) Play(url string) (err error) {
	_, err = p.control.Play(url)
	return
}

func (p *SpotifyPlayer) Pause(pauseState bool) (err error) {
	_, err = p.control.SetPauseState(pauseState)
	return
}

func (p *SpotifyPlayer) Stop() (err error) {
	_, err = p.control.SetPauseState(true)
	return
}
