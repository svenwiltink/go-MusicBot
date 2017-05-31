package songplayer

import (
	"testing"
)

func TestYoutubeSearching(t *testing.T) {
	p, err := NewYoutubePlayer("mpv", "/tmp/.bot-mpv-input")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	items, err := p.Search(SEARCH_TYPE_TRACK, "totalbiscuit", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)

	items, err = p.Search(SEARCH_TYPE_TRACK, "nyan", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)

	items, err = p.Search(SEARCH_TYPE_TRACK, "fail movies", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)
}
