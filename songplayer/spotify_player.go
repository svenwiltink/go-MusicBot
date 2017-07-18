package songplayer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	DEFAULT_AUTHORISE_PORT = 5678
	DEFAULT_AUTHORISE_URL  = "http://musicbot:5678/authorise/"

	MAX_SPOTIFY_PLAYLIST_ITEMS = 1000
)

var ErrNotAuthorised = errors.New("Client has not been authorised yet")

type SpotifyPlayer struct {
	sessionKey     string
	tokenFilePath  string
	playbackDevice string
	client         *spotify.Client
	user           *spotify.PrivateUser
	auth           *spotify.Authenticator
	logger         *logrus.Entry
	authListeners  []func()
}

func NewSpotifyPlayer(spotifyClientID, spotifyClientSecret, tokenFilePath, authoriseRedirectURL string, authoriseHTTPPort int) (p *SpotifyPlayer, authURL string, err error) {
	if authoriseRedirectURL == "" {
		authoriseRedirectURL = DEFAULT_AUTHORISE_URL
	}
	if authoriseHTTPPort <= 0 {
		authoriseHTTPPort = DEFAULT_AUTHORISE_PORT
	}

	auth := spotify.NewAuthenticator(authoriseRedirectURL, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState, spotify.ScopePlaylistReadCollaborative, spotify.ScopePlaylistReadPrivate)
	auth.SetAuthInfo(spotifyClientID, spotifyClientSecret)

	p = &SpotifyPlayer{
		sessionKey:    RandStringBytes(12),
		tokenFilePath: tokenFilePath,
		auth:          &auth,
		logger:        logrus.WithField("songplayer", "Spotify"),
	}
	authURL, err = p.init(authoriseHTTPPort)
	return
}

func (p *SpotifyPlayer) init(authoriseHTTPPort int) (authURL string, err error) {
	// Add our own after authorisation handler
	p.AddAuthorisationListener(p.afterAuthorisation)

	go func() {
		err = http.ListenAndServe(fmt.Sprintf(":%d", authoriseHTTPPort), p)
		if err != nil {
			p.logger.Errorf("SpotifyPlayer.init: Could not start HTTP server on port %d: %v", authoriseHTTPPort, err)
			return
		}
	}()

	token, readErr := p.readToken()
	if readErr == nil {
		client := p.auth.NewClient(token)
		client.AutoRetry = true
		var userErr error
		p.user, userErr = client.CurrentUser()
		if userErr == nil {
			p.logger.Info("SpotifyPlayer.init: Reusing previous spotify token")
			p.client = &client
			p.afterAuthorisation()
			return
		}
		p.logger.Info("SpotifyPlayer.init: Previous token invalid, new authentication needed")
	}

	authURL = p.auth.AuthURL(p.sessionKey)
	p.logger.Infof("SpotifyPlayer.init: Please authorise the MusicBot by visiting the following page in your browser: %s", authURL)
	return
}

func (p *SpotifyPlayer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.client != nil {
		w.WriteHeader(http.StatusPreconditionFailed)
		fmt.Fprint(w, "<h1>Already authenticated!</h1>")
		return
	}

	token, err := p.auth.Token(p.sessionKey, r)
	if err != nil {
		http.Error(w, "Could not get token", http.StatusForbidden)
		p.logger.Warnf("SpotifyPlayer.ServeHTTP: Error pulling token from callback: %v", err)
		return
	}
	if st := r.FormValue("state"); st != p.sessionKey {
		p.logger.Warnf("SpotifyPlayer.ServeHTTP: State mismatch: %v != %v", st, p.sessionKey)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// use the token to get an authenticated client
	client := p.auth.NewClient(token)
	client.AutoRetry = true
	p.client = &client
	p.user, err = p.client.CurrentUser()

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Login completed as %s</h1>", p.user.ID)

	p.writeToken(token)
	for _, l := range p.authListeners {
		l()
	}
}

func (p *SpotifyPlayer) AddAuthorisationListener(listener func()) {
	p.authListeners = append(p.authListeners, listener)
}

func (p *SpotifyPlayer) afterAuthorisation() {
	p.logger = p.logger.WithField("spotifyUser", p.user.ID)

	p.logger.Info("SpotifyPlayer.afterAuthorisation: Successfully authorised")

	// Turn repeat off, as it interferes with the musicplayer
	repErr := p.client.Repeat("off")
	if repErr != nil {
		p.logger.Warnf("SpotifyPlayer.afterAuthorisation: Error setting repeat setting: %v", repErr)
	}

	// Turn shuffle off
	shufErr := p.client.Shuffle(false)
	p.client.Shuffle(false)
	if shufErr != nil {
		p.logger.Warnf("SpotifyPlayer.afterAuthorisation: Error setting shuffle setting: %v", shufErr)
	}
}

func (p *SpotifyPlayer) writeToken(token *oauth2.Token) (err error) {
	if p.tokenFilePath == "" {
		err = errors.New("token filepath invalid")
		return
	}
	buf, err := json.Marshal(token)
	if err != nil {
		p.logger.Warnf("SpotifyPlayer.writeToken: Error marshalling spotify token: %v", err)
		return
	}
	err = ioutil.WriteFile(p.tokenFilePath, buf, 0755)
	if err != nil {
		p.logger.Warnf("SpotifyPlayer.writeToken: Error writing spotify token: %v", err)
		return
	}
	return
}

func (p *SpotifyPlayer) readToken() (token *oauth2.Token, err error) {
	if p.tokenFilePath == "" {
		err = errors.New("token filepath invalid")
		return
	}
	buf, err := ioutil.ReadFile(p.tokenFilePath)
	if err != nil {
		p.logger.Warnf("SpotifyPlayer.readToken: Error reading spotify token: %v", err)
		return
	}
	token = &oauth2.Token{}
	err = json.Unmarshal(buf, token)
	if err != nil {
		p.logger.Warnf("SpotifyPlayer.readToken: Error unmarshalling spotify token: %v", err)
		return
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
	if p.client == nil {
		err = ErrNotAuthorised
		return
	}

	tp, id, userID, err := GetSpotifyTypeAndIDFromURL(url)
	if err != nil {
		p.logger.Errorf("SpotifyPlayer.GetSongs: Could not parse URL [%s] %v", url, err)
		return
	}
	switch tp {
	case TYPE_TRACK:
		var track *spotify.FullTrack
		track, err = p.client.GetTrack(spotify.ID(id))
		if err != nil {
			p.logger.Errorf("SpotifyPlayer.GetSongs: Could not get track data for URL [%s] %v", url, err)
			return
		}

		songs = append(songs,
			NewSong(GetSpotifyTrackName(&track.SimpleTrack), track.TimeDuration(), string(track.URI), GetSpotifyAlbumImage(&track.Album)),
		)
	case TYPE_ALBUM:
		var album *spotify.FullAlbum
		album, err = p.client.GetAlbum(spotify.ID(id))
		if err != nil {
			p.logger.Errorf("SpotifyPlayer.GetSongs: Could not get album data for URL [%s] %v", url, err)
			return
		}
		for _, track := range album.Tracks.Tracks {
			songs = append(songs,
				NewSong(GetSpotifyTrackName(&track), track.TimeDuration(), string(track.URI), GetSpotifyAlbumImage(&album.SimpleAlbum)),
			)
		}
	case TYPE_PLAYLIST:
		offset := 0
		limit := 50
		for {
			var listTracks *spotify.PlaylistTrackPage
			listTracks, err = p.client.GetPlaylistTracksOpt(userID, spotify.ID(id), &spotify.Options{
				Offset: &offset,
				Limit:  &limit,
			}, "")
			if err != nil {
				p.logger.Errorf("SpotifyPlayer.GetSongs: Could not get playlist data for URL [%s] %v", url, err)
				return
			}

			for _, track := range listTracks.Tracks {
				if strings.HasPrefix(string(track.Track.URI), "spotify:local:") {
					p.logger.Infof("SpotifyPlayer.GetSongs: Skipping local song %s", track.Track.URI)
					continue
				}
				songs = append(songs,
					NewSong(GetSpotifyTrackName(&track.Track.SimpleTrack), track.Track.TimeDuration(), string(track.Track.URI), GetSpotifyAlbumImage(&track.Track.Album)),
				)
			}
			offset += limit
			if offset >= listTracks.Total || offset > MAX_SPOTIFY_PLAYLIST_ITEMS {
				break
			}
		}
	}
	return
}

func (p *SpotifyPlayer) Search(searchType SearchType, searchStr string, limit int) (results []PlayableSearchResult, err error) {
	if p.client == nil {
		err = ErrNotAuthorised
		return
	}

	results, err = GetSpotifySearchResults(p.client, searchType, searchStr, limit)
	if err != nil {
		p.logger.Errorf("SpotifyPlayer.Search: Error searching [%d | %s | %d] %v", searchType, searchStr, limit)
		return
	}
	return
}

func (p *SpotifyPlayer) SetPlaybackDevice(playbackDevice string) {
	logrus.Infof("SpotifyPlayer.SetPlaybackDevice: Setting spotify playback device to [%s]", playbackDevice)

	p.playbackDevice = playbackDevice
}

func (p *SpotifyPlayer) setPlaybackDevice() (err error) {
	devices, err := p.client.PlayerDevices()
	if err != nil {
		p.logger.Errorf("SpotifyPlayer.setPlaybackDevice: Error getting devices. %v", err)
		return
	}

	var newDev, oldDev *spotify.PlayerDevice
	for i := range devices {
		if devices[i].Active {
			oldDev = &devices[i]
		}
		if strings.ToLower(devices[i].Name) == strings.ToLower(p.playbackDevice) {
			newDev = &devices[i]
		}
	}

	if newDev == nil {
		err = errors.New("device not found")
		return
	}

	if newDev.Active {
		// Device is already active, nothing to do
		return
	}

	if newDev.Restricted {
		err = errors.New("device is restricted")
		return
	}

	if oldDev != nil {
		p.logger.Infof("SpotifyPlayer.setPlaybackDevice: Transferring playback from %s (%s) to %s (%s)", oldDev.Name, oldDev.Type, newDev.Name, newDev.Type)
	} else {
		p.logger.Infof("SpotifyPlayer.setPlaybackDevice: Transferring playback to %s (%s)", newDev.Name, newDev.Type)
	}

	err = p.client.TransferPlayback(newDev.ID, true)
	if err != nil {
		p.logger.Errorf("SpotifyPlayer.setPlaybackDevice: Error transferring playback to %s (%s). %v", newDev.Name, newDev.Type, err)
	}
	return
}

func (p *SpotifyPlayer) Play(url string) (err error) {
	if p.client == nil {
		err = ErrNotAuthorised
		return
	}

	if p.playbackDevice != "" {
		err = p.setPlaybackDevice()
		if err != nil {
			p.logger.Errorf("SpotifyPlayer.Play: Error setting playback device [%s | %s] %v", p.playbackDevice, url, err)
		}
	}

	URI := spotify.URI(url)
	err = p.client.PlayOpt(&spotify.PlayOptions{
		URIs: []spotify.URI{URI},
	})
	if err != nil {
		p.logger.Errorf("SpotifyPlayer.Play: Error playing [%s] %v", url, err)
		return
	}
	return
}

func (p *SpotifyPlayer) Seek(positionSeconds int) (err error) {
	if p.client == nil {
		err = ErrNotAuthorised
		return
	}

	err = p.client.Seek(positionSeconds * 1000)
	if err != nil {
		p.logger.Errorf("SpotifyPlayer.Seek: Error seeking [%d] %v", positionSeconds, err)
		return
	}
	return
}

func (p *SpotifyPlayer) Pause(pauseState bool) (err error) {
	if p.client == nil {
		err = ErrNotAuthorised
		return
	}

	if pauseState {
		err = p.client.Pause()
		if err != nil {
			p.logger.Errorf("SpotifyPlayer.Pause: Error pausing: %v", err)
			return
		}
		return
	}
	err = p.client.Play()
	if err != nil {
		p.logger.Errorf("SpotifyPlayer.Pause: Error unpausing: %v", err)
		return
	}
	return
}

func (p *SpotifyPlayer) Stop() (err error) {
	if p.client == nil {
		err = ErrNotAuthorised
		return
	}

	err = p.client.Pause()
	if err != nil {
		p.logger.Errorf("SpotifyPlayer.Stop: Error stopping: %v", err)
		return
	}
	return
}
