package songplayer

import "time"

type SongResult struct {
	*Song

	Type SearchType
}

func NewSongResult(tp SearchType, title string, duration time.Duration, URL, imageURL string) *SongResult {
	song := &Song{
		Title:    title,
		Duration: duration,
		URL:      URL,
		ImageURL: imageURL,
	}
	return &SongResult{
		Song: song,
		Type: tp,
	}
}

func (sr *SongResult) GetType() SearchType {
	return sr.Type
}
