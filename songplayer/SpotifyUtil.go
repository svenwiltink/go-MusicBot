package songplayer

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"github.com/zmb3/spotify"
)

type Type int

const (
	TYPE_TRACK Type = 1 + iota
	TYPE_ALBUM
	TYPE_PLAYLIST
)

func GetSpotifyTrackName(track spotify.SimpleTrack) (name string) {
	name = track.Name
	var artistNames []string
	for _, artist := range track.Artists {
		artistNames = append(artistNames, artist.Name)
	}
	if len(artistNames) > 0 {
		name = fmt.Sprintf("%s - %s", track.Name, strings.Join(artistNames, ", "))
	}
	return
}

func GetSpotifyTrackImage(album spotify.SimpleAlbum) (imageURL string) {
	if len (album.Images) > 0 {
		imageURL = album.Images[0].URL
	}
	return
}

func GetSpotifyTypeAndIDFromURL(url string) (tp Type, id, userID string, err error) {
	lowerURL := strings.ToLower(url)
	if strings.Contains(lowerURL, "spotify.com") {
		var idPos int
		switch {
		case strings.Contains(lowerURL, "/track/"):
			//Handle: https://open.spotify.com/track/4uLU6hMCjMI75M1A2tKUQC
			tp = TYPE_TRACK
			idPos = strings.LastIndex(lowerURL, "/track/") + len("/track/")
		case strings.Contains(lowerURL, "/album/"):
			// Handle: https://open.spotify.com/album/4vSfHrq6XxVyMcJ6PguFR2
			tp = TYPE_ALBUM
			idPos = strings.LastIndex(lowerURL, "/album/") + len("/album/")
		case strings.Contains(lowerURL, "/playlist/"):
			// Handle: https://open.spotify.com/user/tana.cross/playlist/2xLFotd9GVVQ6Jde7B3i3B
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

	var idPos int
	switch {
	case strings.Contains(lowerURL, ":track:"):
		// Handle: spotify:track:2cBGl1Ehr1D9xbqNmraqb4
		tp = TYPE_TRACK
		idPos = strings.LastIndex(lowerURL, ":track:") + len(":track:")
	case strings.Contains(lowerURL, ":album:"):
		// Handle: spotify:album:35LnYSwPbgGPQSXNTjpOO8
		tp = TYPE_ALBUM
		idPos = strings.LastIndex(lowerURL, ":album:") + len(":album:")
	case strings.Contains(lowerURL, ":playlist:"):
		// Handle: spotify:user:111208973:playlist:4XGuyS11n99eMqe1OvN8jq
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ11234567890"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
