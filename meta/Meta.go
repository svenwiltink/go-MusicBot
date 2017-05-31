package meta

import (
	"time"
)

type Meta struct {
	Identifier string
	Title      string
	Artist     string
	Album      string
	URL        string
	Duration   time.Duration
	ImageURL   string
}

func (m *Meta) GetTitle() string {
	return m.Title
}

func (m *Meta) GetDuration() time.Duration {
	return m.Duration
}

func (m *Meta) GetURL() string {
	return m.URL
}
