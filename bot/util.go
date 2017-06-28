package bot

import (
	"errors"
	"fmt"
	"github.com/svenwiltink/go-musicbot/player"
	"github.com/svenwiltink/go-musicbot/songplayer"
	"github.com/svenwiltink/go-musicbot/util"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	NORMAL_CHARACTER    = string('\u000F')
	BOLD_CHARACTER      = string('\u0002')
	UNDERLINE_CHARACTER = string('\u001F')
	ITALIC_CHARACTER    = string('\u001D')
	INVERSE_CHARACTER   = string('\u0016')
	COLOUR_CHARACTER    = string('\u0003')
)

type Color string

const (
	COLOUR_WHITE       Color = "00"
	COLOUR_BLACK       Color = "01"
	COLOUR_BLUE        Color = "02"
	COLOUR_GREEN       Color = "03"
	COLOUR_RED         Color = "04"
	COLOUR_BROWN       Color = "05"
	COLOUR_PURPLE      Color = "06"
	COLOUR_ORANGE      Color = "07"
	COLOUR_YELLOW      Color = "08"
	COLOUR_LIGHT_GREEN Color = "09"
	COLOUR_TEAL        Color = "10"
	COLOUR_CYAN        Color = "11"
	COLOUR_LIGHT_BLUE  Color = "12"
	COLOUR_PINK        Color = "13"
	COLOUR_GREY        Color = "14"
	COLOUR_LIGHT_GREY  Color = "15"
	COLOUR_DEFAULT     Color = "99"
	COLOUR_NONE        Color = ""
)

func boldText(s string) (bs string) {
	return BOLD_CHARACTER + s + BOLD_CHARACTER
}

func underlineText(s string) (us string) {
	return UNDERLINE_CHARACTER + s + UNDERLINE_CHARACTER
}

func italicText(s string) (us string) {
	return ITALIC_CHARACTER + s + ITALIC_CHARACTER
}

func inverseText(s string) (is string) {
	return INVERSE_CHARACTER + s + INVERSE_CHARACTER
}

func getColourCode(foreground, background Color) (code string) {
	if foreground == COLOUR_NONE {
		return
	}
	if background == COLOUR_NONE {
		code = fmt.Sprintf("%s%s", COLOUR_CHARACTER, foreground)
		return
	}
	code = fmt.Sprintf("%s%s,%s", COLOUR_CHARACTER, foreground, background)
	return
}

func colourText(foreground, backgroundColor Color, s string) (cs string) {
	cc := getColourCode(foreground, backgroundColor)
	endCode := getColourCode(COLOUR_DEFAULT, COLOUR_DEFAULT)
	cs = cc + s + endCode
	return
}

func formatSong(song songplayer.Playable) (s string) {
	s = fmt.Sprintf("%s %s%s%s", song.GetTitle(), getColourCode(COLOUR_TEAL, COLOUR_NONE), util.FormatDuration(song.GetDuration()), getColourCode(COLOUR_DEFAULT, COLOUR_NONE))
	return
}

func progressString(total, remaining time.Duration) (progress string) {
	percent := int(float64(total-remaining) / float64(total) * 100)

	progress = getColourCode(COLOUR_BLUE, COLOUR_BLUE)
	for i := 0; i < percent; i++ {
		progress += "#"
	}
	progress += "#" + getColourCode(COLOUR_LIGHT_GREY, COLOUR_LIGHT_GREY)
	if percent < 100 {
		for i := percent + 1; i < 100; i++ {
			progress += "-"
		}
	}
	progress += getColourCode(COLOUR_DEFAULT, COLOUR_NONE)
	return
}

func getPlayerNames(player player.MusicPlayer) (names []string) {
	for _, playr := range player.GetSongPlayers() {
		names = append(names, playr.Name())
	}
	return
}

func searchSongs(player player.MusicPlayer, parameters []string, limit int) (results map[string][]songplayer.PlayableSearchResult, err error) {
	results = make(map[string][]songplayer.PlayableSearchResult)

	if len(parameters) < 1 {
		err = errors.New("Missing search parameter")
		return
	}

	searchType := songplayer.SEARCH_TYPE_TRACK
	ok, searchType := songplayer.GetSearchType(parameters[0])
	if ok {
		parameters = parameters[1:] // cut off the searchType parameter
		if len(parameters) < 1 {
			err = errors.New("Missing search parameter")
			return
		}
	}

	plyr := player.GetSongPlayer(parameters[0])
	if plyr != nil {
		parameters = parameters[1:] // cut off the player parameter
		if len(parameters) < 1 {
			err = errors.New("Missing search parameter")
			return
		}
	}

	searchFunc := func(songPlayer songplayer.SongPlayer, searchStr string) (err error) {
		var items []songplayer.PlayableSearchResult
		items, err = songPlayer.Search(searchType, searchStr, limit)
		if err != nil {
			return
		}
		for _, item := range items {
			results[songPlayer.Name()] = append(results[songPlayer.Name()], item)
		}
		return
	}

	if plyr != nil {
		searchErr := searchFunc(plyr, strings.Join(parameters, " "))
		if searchErr != nil {
			logrus.Errorf("MusicBot.searchSongs: Error searching [%s | %d | %v] %v", plyr.Name(), searchType, parameters, searchErr)
		}
		return
	}

	for _, songPlayer := range player.GetSongPlayers() {
		searchErr := searchFunc(songPlayer, strings.Join(parameters, " "))
		if searchErr != nil {
			logrus.Errorf("MusicBot.searchSongs: Error searching [%s | %d | %v] %v", songPlayer.Name(), searchType, parameters, searchErr)
		}
	}
	return
}

func GetMusicBotStringFormatted() (s string) {
	s = fmt.Sprintf("%s%s%sM%su%ss%si%sc%sB%so%st%s%s%s",
		BOLD_CHARACTER,
		UNDERLINE_CHARACTER,
		getColourCode(COLOUR_GREEN, COLOUR_NONE),
		getColourCode(COLOUR_RED, COLOUR_NONE),
		getColourCode(COLOUR_PURPLE, COLOUR_NONE),
		getColourCode(COLOUR_ORANGE, COLOUR_NONE),
		getColourCode(COLOUR_LIGHT_GREEN, COLOUR_NONE),
		getColourCode(COLOUR_CYAN, COLOUR_NONE),
		getColourCode(COLOUR_LIGHT_BLUE, COLOUR_NONE),
		getColourCode(COLOUR_PINK, COLOUR_NONE),
		getColourCode(COLOUR_DEFAULT, COLOUR_NONE),
		BOLD_CHARACTER,
		UNDERLINE_CHARACTER,
	)
	return
}
