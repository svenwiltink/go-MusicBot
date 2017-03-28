package player

import "time"

type ListItem struct {
	title    string
	duration time.Duration
	url      string
}

func NewListItem(title string, duration time.Duration, URL string) *ListItem {
	return &ListItem{
		title:    title,
		duration: duration,
		url:      URL,
	}
}

func (i *ListItem) Title() string {
	return i.title
}

func (i *ListItem) Duration() time.Duration {
	return i.duration
}

func (i *ListItem) URL() string {
	return i.url
}
