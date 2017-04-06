package api

import (
	"gitlab.transip.us/swiltink/go-MusicBot/playlist"
	"gitlab.transip.us/swiltink/go-MusicBot/util"
	"time"
)

type Item struct {
	Title            string
	Seconds          int
	SecondsRemaining int
	FormattedTime    string
	URL              string
}

type Status struct {
	Status  playlist.Status
	Current *Item
	List    []Item
}

type Event struct {
	Event     string
	Arguments []interface{}
}

type Command struct {
	Command   string
	Arguments []string
}

type CommandResponse struct {
	Command string
	Success bool
	Error   string
	Status  *Status `json:",omitempty"`
}

func getAPIItem(itm playlist.ItemInterface, remaining time.Duration) (newItem *Item) {
	if itm != nil {
		duration := itm.GetDuration()

		newItem = &Item{
			Title:            itm.GetTitle(),
			URL:              itm.GetURL(),
			Seconds:          int(duration.Seconds()),
			SecondsRemaining: int(remaining.Seconds()),
			FormattedTime:    util.FormatSongLength(duration),
		}
	}
	return
}

func getAPIItems(itms []playlist.ItemInterface) (newItems []Item) {
	for _, itm := range itms {
		if itm == nil {
			continue
		}
		newItems = append(newItems, *getAPIItem(itm, itm.GetDuration()))
	}
	return
}

func getCommandResponse(cmd *Command, err error) (resp CommandResponse) {
	resp.Command = cmd.Command
	resp.Success = err == nil
	resp.Error = err.Error()
	return
}
