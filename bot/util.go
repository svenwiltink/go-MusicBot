package bot

import (
	"fmt"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"gitlab.transip.us/swiltink/go-MusicBot/songplayer"
	"gitlab.transip.us/swiltink/go-MusicBot/util"
	"strings"
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
	s = fmt.Sprintf("%s %s%s%s", song.GetTitle(), getColourCode(COLOUR_TEAL, COLOUR_NONE), util.FormatSongLength(song.GetDuration()), getColourCode(COLOUR_DEFAULT, COLOUR_NONE))
	return
}

func searchSongs(player player.MusicPlayer, parameters []string) (results map[string][]songplayer.Playable, err error) {
	results = make(map[string][]songplayer.Playable)

	searchFunc := func(songPlayer songplayer.SongPlayer, searchStr string) {
		var items []songplayer.Playable
		items, err = songPlayer.SearchSongs(searchStr, 3)
		if err != nil {
			return
		}
		for _, item := range items {
			results[songPlayer.Name()] = append(results[songPlayer.Name()], item)
		}
	}

	plyr := player.GetSongPlayer(parameters[0])
	if plyr != nil {
		searchFunc(plyr, strings.Join(parameters[1:], " "))
		return
	}

	for _, songPlayer := range player.GetSongPlayers() {
		searchFunc(songPlayer, strings.Join(parameters, " "))
	}
	return
}
