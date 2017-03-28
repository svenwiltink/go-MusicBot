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
