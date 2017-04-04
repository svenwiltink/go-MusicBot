package songplayer

import (
	"testing"
)

func TestYoutubeSearching(t *testing.T) {
	p, err := NewYoutubePlayer()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	items, err := p.SearchSongs("totalbiscuit", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)

	items, err = p.SearchSongs("nyan", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)

	items, err = p.SearchSongs("fail movies", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)
}
