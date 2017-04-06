package player

import "time"

type ListItem struct {
	Title    string
	Duration time.Duration
	URL      string
}

func NewListItem(title string, duration time.Duration, URL string) *ListItem {
	return &ListItem{
		Title:    title,
		Duration: duration,
		URL:      URL,
	}
}

func (i *ListItem) GetTitle() string {
	return i.Title
}

func (i *ListItem) GetDuration() time.Duration {
	return i.Duration
}

func (i *ListItem) GetURL() string {
	return i.URL
}
