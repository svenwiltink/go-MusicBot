package player

import (
	"time"
)

type Stats struct {
	TotalTimePlayed     time.Duration
	TotalSongsPlayed    int
	TotalSongsQueued    int
	TotalTimesNext      int
	TotalTimesPrevious  int
	TotalTimesPaused    int
	SongsPlayedByPlayer map[string]int
}
