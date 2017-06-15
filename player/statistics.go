package player

import (
	"time"
)

type Statistics struct {
	TotalTimePlayed     time.Duration
	TotalSongsPlayed    int
	TotalSongsQueued    int
	TotalTimesNext      int
	TotalTimesPrevious  int
	TotalTimesJump      int
	TotalTimesPaused    int
	TimeByPlayer        map[string]time.Duration
	SongsPlayedByPlayer map[string]int
	SongsAddedByUser    map[string]int
}
