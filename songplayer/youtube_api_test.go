package songplayer

import (
	"testing"
)

func TestYoutubeAPISearchPlaylist(t *testing.T) {
	yt := NewYoutubeAPI()

	items, err := yt.Search(SEARCH_TYPE_PLAYLIST, "cooptional", 10)
	if err != nil {
		t.Fail()
		t.Logf("Error: %v", err)
	}

	t.Logf("Results: ")
	for _, p := range items {
		t.Logf("%s | %s | %d", p.GetTitle(), p.GetURL(), p.GetType())
	}

}
