package player

import (
	"github.com/SvenWiltink/go-MusicBot/songplayer"
	"testing"
	"time"
)

func TestPlay(t *testing.T) {
	s, err := songplayer.NewSpotifyPlayer()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	p := NewPlayer()
	p.AddSongPlayer(s)

	items, err := p.AddSongs("spotify:album:3fa5cl6Nplripk1h9z1SFv")
	if len(items) != 8 || err != nil {
		t.Log(items)
		t.Log(err)
		t.Fail()
	}

	err = p.Play()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	time.Sleep(time.Second * 10)

	err = p.Pause()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	time.Sleep(time.Second * 4)

	err = p.Pause()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	time.Sleep(time.Second * 10)

	item, err := p.Next()
	if item == nil || err != nil {
		t.Log(item, err)
		t.Fail()
	}
}
