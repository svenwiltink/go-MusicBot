package songplayer

import (
	"testing"
	"time"
)

func TestSpotifyConnectSearching(t *testing.T) {
	p, authURL, err := NewSpotifyConnectPlayer("5b83b8fcf4c142aba7f84ee7985a45c9", "<SECRIT>", "", 0)
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
}
