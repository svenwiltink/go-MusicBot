package songplayer

import (
	"github.com/zmb3/spotify"
	"log"
	"testing"
	"time"
)

func TestSpotifyConnectSearching(t *testing.T) {
	p, authURL, err := NewSpotifyConnectPlayer("5b83b8fcf4c142aba7f84ee7985a45c9", "<secrit>", "", 0)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Logf("AuthURL: %s", authURL)

	done := false
	p.AddAuthorisationListener(func() {
		done = true
	})

	for {
		if done {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	items, err := p.SearchSongs("green day boulevard", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)

	items, err = p.SearchSongs("adele chasing pavement", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)

	items, err = p.SearchSongs("hallelujah", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)

	tp, id, uid, err := GetSpotifyTypeAndIDFromURL("spotify:user:111208973:playlist:4XGuyS11n99eMqe1OvN8jq")
	if tp != TYPE_PLAYLIST || id != "4XGuyS11n99eMqe1OvN8jq" || err != nil || uid != "111208973" {
		t.Log(string(tp), id, uid, err)
		t.Fail()
	}

	tracks, err := p.client.GetPlaylistTracks(uid, spotify.ID(id))
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	for _, t := range tracks.Tracks {
		log.Println(t.Track.Name)
	}
}
