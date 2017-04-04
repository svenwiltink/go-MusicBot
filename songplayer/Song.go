package songplayer

import "time"

type Song struct {
	Title    string
	Duration time.Duration
	URL      string
}

func NewSong(title string, duration time.Duration, URL string) *Song {
	return &Song{
		Title:    title,
		Duration: duration,
		URL:      URL,
	}
}

func (i *Song) GetTitle() string {
	return i.Title
}

func (i *Song) GetDuration() time.Duration {
	return i.Duration
}

func (i *Song) GetURL() string {
	return i.URL
}
