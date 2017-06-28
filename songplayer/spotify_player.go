package songplayer

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/vansante/go-spotify-control"
	"github.com/zmb3/spotify"
	"time"
)

type SpotifyPlayer struct {
	host    string
	control *spotifycontrol.SpotifyControl
}

func NewSpotifyPlayer(host string) (p *SpotifyPlayer, err error) {
	cntrl, err := spotifycontrol.NewSpotifyControl(host, 1*time.Second)
	if err != nil {
		return
	}

	p = &SpotifyPlayer{
		host:    host,
		control: cntrl,
	}
	return
}

func (p *SpotifyPlayer) Name() (name string) {
	return "Spotify"
}

func (p *SpotifyPlayer) CanPlay(url string) (canPlay bool) {
	_, id, _, err := GetSpotifyTypeAndIDFromURL(url)
	canPlay = err == nil && id != ""
	return
}

func (p *SpotifyPlayer) GetSongs(url string) (songs []Playable, err error) {
	tp, id, _, err := GetSpotifyTypeAndIDFromURL(url)
	if err != nil {
		logrus.Errorf("SpotifyPlayer.GetSongs: Could not parse URL [%s] %v", url, err)
		return
	}
	switch tp {
	case TYPE_TRACK:
		var track *spotify.FullTrack
		track, err = spotify.DefaultClient.GetTrack(spotify.ID(id))
		if err != nil {
			logrus.Errorf("SpotifyPlayer.GetSongs: Could not get track data for URL [%s] %v", url, err)
			return
		}

		songs = append(songs,
			NewSong(GetSpotifyTrackName(&track.SimpleTrack), track.TimeDuration(), string(track.URI), GetSpotifyAlbumImage(&track.Album)),
		)
	case TYPE_ALBUM:
		var album *spotify.FullAlbum
		album, err = spotify.DefaultClient.GetAlbum(spotify.ID(id))
		if err != nil {
			logrus.Errorf("SpotifyPlayer.GetSongs: Could not get album data for URL [%s] %v", url, err)
			return
		}
		for _, track := range album.Tracks.Tracks {
			songs = append(songs,
				NewSong(GetSpotifyTrackName(&track), track.TimeDuration(), string(track.URI), GetSpotifyAlbumImage(&album.SimpleAlbum)),
			)
		}
	case TYPE_PLAYLIST:
		err = errors.New("Playlists are not supported yet, they require oauth :(")
		return
	}
	return
}

func (p *SpotifyPlayer) Search(searchType SearchType, searchStr string, limit int) (results []PlayableSearchResult, err error) {
	results, err = GetSpotifySearchResults(spotify.DefaultClient, searchType, searchStr, limit)
	return
}

func (p *SpotifyPlayer) Play(url string) (err error) {
	_, err = p.control.Play(url)
	p.restartAndRetry(err, func() {
		_, err = p.control.Play(url)
	})
	if err != nil {
		logrus.Errorf("SpotifyPlayer.Play: Could not play URL [%s] %v", url, err)
		return
	}
	return
}

func (p *SpotifyPlayer) Seek(positionSeconds int) (err error) {
	err = errors.New("seek is not supported")
	return
}

func (p *SpotifyPlayer) Pause(pauseState bool) (err error) {
	_, err = p.control.SetPauseState(pauseState)
	p.restartAndRetry(err, func() {
		_, err = p.control.SetPauseState(pauseState)
	})
	if err != nil {
		logrus.Errorf("SpotifyPlayer.Pause: Could not pause [%v] %v", pauseState, err)
		return
	}
	return
}

func (p *SpotifyPlayer) Stop() (err error) {
	_, err = p.control.SetPauseState(true)
	p.restartAndRetry(err, func() {
		_, err = p.control.SetPauseState(true)
	})
	if err != nil {
		logrus.Errorf("SpotifyPlayer.Pause: Could not stop: %v", err)
		return
	}
	return
}

func (p *SpotifyPlayer) restartAndRetry(spErr error, retryFunc func()) (err error) {
	if spErr == nil {
		return
	}
	logrus.Infof("SpotifyPlayer.restartAndRetry: Error encountered, restarting control to try again. Error: %v", spErr)

	var control *spotifycontrol.SpotifyControl
	control, err = spotifycontrol.NewSpotifyControl(p.host, 1*time.Second)
	if err != nil {
		logrus.Errorf("SpotifyPlayer.restartAndRetry: Restart unsuccessful: %v", err)
		err = spErr
		return
	}
	p.control = control
	logrus.Info("SpotifyPlayer.restartAndRetry: Restart finished, retrying")
	retryFunc()
	return
}
