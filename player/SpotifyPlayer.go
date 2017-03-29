package player

import (
	"errors"
	"fmt"
	"github.com/vansante/go-spotify-control"
	"github.com/zmb3/spotify"
	"strings"
	"time"
)

type Type int

const (
	TYPE_TRACK Type = 1 + iota
	TYPE_ALBUM
	TYPE_PLAYLIST
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
	_, id, err := p.getTypeAndIDFromURL(url)
	canPlay = err == nil && id != ""
	return
}

func (p *SpotifyPlayer) GetItems(url string) (items []ListItem, err error) {
	tp, id, err := p.getTypeAndIDFromURL(url)
	var tracks []spotify.SimpleTrack
	switch tp {
	case TYPE_TRACK:
		var track *spotify.FullTrack
		track, err = spotify.DefaultClient.GetTrack(spotify.ID(id))
		if err != nil {
			err = fmt.Errorf("[SpotifyPlayer] Could not get track meta for url: %v", err)
			return
		}
		tracks = append(tracks, track.SimpleTrack)
	case TYPE_ALBUM:
		var album *spotify.FullAlbum
		album, err = spotify.DefaultClient.GetAlbum(spotify.ID(id))
		if err != nil {
			err = fmt.Errorf("[SpotifyPlayer] Could not get album meta for url: %v", err)
			return
		}
		for _, track := range album.Tracks.Tracks {
			tracks = append(tracks, track)
		}
	case TYPE_PLAYLIST:
		err = errors.New("Playlists are not supported yet, they require tokens :(")
		return
	}

	for _, track := range tracks {
		name := track.Name
		if len(track.Artists) > 0 {
			name = fmt.Sprintf("%s - %s", track.Name, track.Artists[0].Name)
		}
		items = append(items, *NewListItem(name, track.TimeDuration(), url))
	}
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

func (p *SpotifyPlayer) getTypeAndIDFromURL(url string) (tp Type, id string, err error) {
	lowerURL := strings.ToLower(url)
	if strings.Contains(lowerURL, "spotify.com") {
		var position int
		switch {
		//https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC
		case strings.Contains(lowerURL, "/track/"):
			tp = TYPE_TRACK
			position = strings.LastIndex(lowerURL, "/track/") + len("/track/")
		case strings.Contains(lowerURL, "/album/"):
			tp = TYPE_ALBUM
			position = strings.LastIndex(lowerURL, "/album/") + len("/album/")
		case strings.Contains(lowerURL, "/playlist/"):
			//https://open.spotify.com/user/tana.cross/playlist/2xLFotd9GVVQ6Jde7B3i3B
			tp = TYPE_PLAYLIST
			position = strings.LastIndex(lowerURL, "/playlist/") + len("/playlist/")
		default:
			err = fmt.Errorf("Unknown spotify url: %s", url)
			return
		}
		id = url[:position]
		return
	}

	// spotify:track:2cBGl1Ehr1D9xbqNmraqb4
	var position int
	switch {
	case strings.Contains(lowerURL, ":track:"):
		tp = TYPE_TRACK
		position = strings.LastIndex(lowerURL, ":track:") + len(":track:")
	case strings.Contains(lowerURL, ":album:"):
		tp = TYPE_ALBUM
		position = strings.LastIndex(lowerURL, ":album:") + len(":album:")
	case strings.Contains(lowerURL, ":playlist:"):
		tp = TYPE_PLAYLIST
		position = strings.LastIndex(lowerURL, ":playlist:") + len(":playlist:")
	default:
		err = fmt.Errorf("Unknown spotify url: %s", url)
		return
	}
	id = url[:position]
	return
}
