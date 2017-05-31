package songplayer

import "time"

type Song struct {
	Title    string
	Duration time.Duration
	URL      string
	ImageURL string
}

func NewSong(title string, duration time.Duration, URL, imageURL string) *Song {
	return &Song{
		Title:    title,
		Duration: duration,
		URL:      URL,
		ImageURL: imageURL,
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

func (i *Song) GetImageURL() string {
	return i.ImageURL
}
