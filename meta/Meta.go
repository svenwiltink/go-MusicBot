package meta

import (
	"time"
)

type Meta struct {
	Identifier string
	Title      string
	Artist     string
	Album      string
	Source     string
	Duration   time.Duration
}

func (m *Meta) GetTitle() string {
	return m.Title
}

func (m *Meta) GetDuration() time.Duration {
	return m.Duration
}

func (m *Meta) GetURL() string {
	return m.Source
}
