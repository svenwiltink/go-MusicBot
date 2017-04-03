package util

import (
	"fmt"
	"time"
)

func FormatSongLength(duration time.Duration) (form string) {
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) - (minutes * 60)

	if minutes > 60 {
		hours := minutes / 60
		minutes -= hours * 60
		form = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
		return
	}
	form = fmt.Sprintf("%02d:%02d", minutes, seconds)
	return
}
