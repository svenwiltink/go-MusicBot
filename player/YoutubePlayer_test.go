package player

import (
	"testing"
)

func TestYoutubeSearching(t *testing.T) {
	p, err := NewYoutubePlayer()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	items, err := p.SearchItems("totalbiscuit", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)

	items, err = p.SearchItems("nyan", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)

	items, err = p.SearchItems("fail movies", 3)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	t.Log("Findings: ", items)
}
