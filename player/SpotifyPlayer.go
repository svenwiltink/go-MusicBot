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
	_, id, _, err := p.getTypeAndIDFromURL(url)
	canPlay = err == nil && id != ""
	return
}

func (p *SpotifyPlayer) GetItems(url string) (items []ListItem, err error) {
	tp, id, _, err := p.getTypeAndIDFromURL(url)
	var tracks []spotify.SimpleTrack
	switch tp {
	case TYPE_TRACK:
		var track *spotify.FullTrack
		track, err = spotify.DefaultClient.GetTrack(spotify.ID(id))
		if err != nil {
			err = fmt.Errorf("[SpotifyPlayer] Could not get track meta for URL: %v", err)
			return
		}
		tracks = append(tracks, track.SimpleTrack)
	case TYPE_ALBUM:
		var album *spotify.FullAlbum
		album, err = spotify.DefaultClient.GetAlbum(spotify.ID(id))
		if err != nil {
			err = fmt.Errorf("[SpotifyPlayer] Could not get album meta for URL: %v", err)
			return
		}
		for _, track := range album.Tracks.Tracks {
			tracks = append(tracks, track)
		}
	case TYPE_PLAYLIST:
		err = errors.New("Playlists are not supported yet, they require oauth :(")
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

func (p *SpotifyPlayer) getTypeAndIDFromURL(url string) (tp Type, id, userID string, err error) {
	lowerURL := strings.ToLower(url)
	if strings.Contains(lowerURL, "spotify.com") {
		var idPos int
		switch {
		//https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC
		case strings.Contains(lowerURL, "/track/"):
			tp = TYPE_TRACK
			idPos = strings.LastIndex(lowerURL, "/track/") + len("/track/")
		case strings.Contains(lowerURL, "/album/"):
			tp = TYPE_ALBUM
			idPos = strings.LastIndex(lowerURL, "/album/") + len("/album/")
		case strings.Contains(lowerURL, "/playlist/"):
			//https://open.spotify.com/user/tana.cross/playlist/2xLFotd9GVVQ6Jde7B3i3B
			tp = TYPE_PLAYLIST
			idPos = strings.LastIndex(lowerURL, "/playlist/")
			uidPos := strings.LastIndex(lowerURL, "/user/") + len("/user/")
			if uidPos >= idPos {
				err = errors.New("Invalid spotify URL format")
				return
			}
			userID = url[uidPos:idPos]

			idPos += len("/playlist/")
		default:
			err = fmt.Errorf("Unknown spotify URL format: %s", url)
			return
		}
		id = url[idPos:]
		return
	}

	// spotify:track:2cBGl1Ehr1D9xbqNmraqb4
	var idPos int
	switch {
	case strings.Contains(lowerURL, ":track:"):
		tp = TYPE_TRACK
		idPos = strings.LastIndex(lowerURL, ":track:") + len(":track:")
	case strings.Contains(lowerURL, ":album:"):
		tp = TYPE_ALBUM
		idPos = strings.LastIndex(lowerURL, ":album:") + len(":album:")
	case strings.Contains(lowerURL, ":playlist:"):
		// spotify:user:111208973:playlist:4XGuyS11n99eMqe1OvN8jq
		tp = TYPE_PLAYLIST
		idPos = strings.LastIndex(lowerURL, ":playlist:")
		uidPos := strings.LastIndex(lowerURL, ":user:") + len(":user:")
		if uidPos >= idPos {
			err = errors.New("Invalid spotify URL format")
			return
		}
		userID = url[uidPos:idPos]

		idPos += len(":playlist:")
	default:
		err = fmt.Errorf("Unknown spotify URL format: %s", url)
		return
	}
	id = url[idPos:]
	return
}
