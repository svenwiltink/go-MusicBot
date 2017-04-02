package util

import (
	"fmt"
	"time"
)

func FormatSongLength(duration time.Duration) (form string) {
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) - (minutes * 60)

	form = fmt.Sprintf("%02d:%02d", minutes, seconds)
	return
}
