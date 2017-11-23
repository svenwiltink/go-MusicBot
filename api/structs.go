package api

import (
	"github.com/svenwiltink/go-musicbot/player"
	"github.com/svenwiltink/go-musicbot/songplayer"
	"github.com/svenwiltink/go-musicbot/util"
	"time"
)

type Version struct {
	VersionTag string
	GitCommit  string
	BuildDate  string
	BuildHost  string
	GoVersion  string
}

type Song struct {
	Title            string
	Seconds          int
	SecondsRemaining int
	FormattedTime    string
	URL              string
	ImageURL         string
}

type Status struct {
	Status  player.Status
	Current *Song
	List    []Song
}

type Command struct {
	Command   string
	Arguments []string
}

type Response struct {
	Type      string
	Timestamp string
	Data      interface{} `json:",omitempty"`
}

type EventResponse struct {
	Response

	Event string
}

type CommandResponse struct {
	Response

	Command string
	Success bool
	Error   string
}

func getAPISong(song songplayer.Playable, remaining time.Duration) (apiSong *Song) {
	if song != nil {
		duration := song.GetDuration()

		apiSong = &Song{
			Title:            song.GetTitle(),
			URL:              song.GetURL(),
			Seconds:          int(duration.Seconds()),
			SecondsRemaining: int(remaining.Seconds()),
			FormattedTime:    util.FormatDuration(duration),
			ImageURL:         song.GetImageURL(),
		}
	}
	return
}

func getAPISongs(songs []songplayer.Playable) (apiSongs []Song) {
	apiSongs = make([]Song, 0)
	for _, song := range songs {
		if song == nil {
			continue
		}
		apiSongs = append(apiSongs, *getAPISong(song, song.GetDuration()))
	}
	return
}

func getCommandResponse(cmd *Command, err error) (resp CommandResponse) {
	resp.Type = "Command"
	resp.Command = cmd.Command
	resp.Success = err == nil
	if err != nil {
		resp.Error = err.Error()
	}
	return
}
