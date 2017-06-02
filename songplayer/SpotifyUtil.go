package songplayer

import (
	"errors"
	"fmt"
	"github.com/zmb3/spotify"
	"math/rand"
	"strings"
	"time"
)

type Type int

const (
	TYPE_TRACK Type = 1 + iota
	TYPE_ALBUM
	TYPE_PLAYLIST
)

func GetSpotifyTrackName(track *spotify.SimpleTrack) (name string) {
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

func GetSpotifyAlbumImage(album *spotify.SimpleAlbum) (imageURL string) {
	if len(album.Images) > 0 {
		imageURL = album.Images[0].URL
	}
	return
}

func GetSpotifyAlbumName(album *spotify.FullAlbum) (name string) {
	name = album.Name
	var artistNames []string
	for _, artist := range album.Artists {
		artistNames = append(artistNames, artist.Name)
	}
	if len(artistNames) > 0 {
		name = fmt.Sprintf("%s - %s", album.Name, strings.Join(artistNames, ", "))
	}
	return
}

func GetSpotifySearchResults(spClient *spotify.Client, searchType SearchType, searchStr string, limit int) (results []PlayableSearchResult, err error) {
	spotifySearchType := spotify.SearchType(spotify.SearchTypeTrack)
	switch searchType {
	case SEARCH_TYPE_ALBUM:
		spotifySearchType = spotify.SearchTypeAlbum
	case SEARCH_TYPE_PLAYLIST:
		spotifySearchType = spotify.SearchTypePlaylist
	}

	spResults, err := spClient.SearchOpt(searchStr, spotifySearchType, &spotify.Options{
		Limit: &limit,
	})

	if spResults.Tracks != nil {
		for _, track := range spResults.Tracks.Tracks {
			results = append(results, NewSongResult(SEARCH_TYPE_TRACK, GetSpotifyTrackName(&track.SimpleTrack), track.TimeDuration(), string(track.URI), GetSpotifyAlbumImage(&track.Album)))
		}
	}

	if spResults.Albums != nil {
		for _, searchAlbum := range spResults.Albums.Albums {
			var album *spotify.FullAlbum
			album, err = spClient.GetAlbum(searchAlbum.ID)
			if err != nil {
				err = fmt.Errorf("[SpotifyConnectPlayer] Could not get album for URL: %v", err)
				return
			}
			var duration time.Duration
			for _, track := range album.Tracks.Tracks {
				duration += track.TimeDuration()
			}
			results = append(results, NewSongResult(SEARCH_TYPE_ALBUM, GetSpotifyAlbumName(album), duration, string(album.URI), GetSpotifyAlbumImage(&album.SimpleAlbum)))
		}
	}

	if spResults.Playlists != nil {
		for _, searchPlaylist := range spResults.Playlists.Playlists {
			var duration time.Duration
			var listTracks *spotify.PlaylistTrackPage
			listTracks, err = spClient.GetPlaylistTracks(searchPlaylist.Owner.ID, searchPlaylist.ID)
			if err != nil {
				err = fmt.Errorf("[SpotifyConnectPlayer] Could not get playlist tracks for URL: %v", err)
				return
			}
			for _, track := range listTracks.Tracks {
				duration += track.Track.TimeDuration()
			}

			imageURL := ""
			if len(searchPlaylist.Images) > 0 {
				imageURL = searchPlaylist.Images[0].URL
			}
			results = append(results, NewSongResult(SEARCH_TYPE_ALBUM, searchPlaylist.Name, duration, string(searchPlaylist.URI), imageURL))
		}
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
