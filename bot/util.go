package bot

import (
	"fmt"
	"gitlab.transip.us/swiltink/go-MusicBot/playlist"
	"gitlab.transip.us/swiltink/go-MusicBot/util"
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
	code = fmt.Sprintf("%s%s,%s", COLOUR_CHARACTER, foreground, background)
	return
}

func colourText(foreground, backgroundColor Color, s string) (cs string) {
	cc := getColourCode(foreground, backgroundColor)
	endCode := getColourCode(COLOUR_DEFAULT, COLOUR_DEFAULT)
	cs = cc + s + endCode
	return
}

func formatSong(song playlist.ItemInterface) (s string) {
	s = fmt.Sprintf("%s %s%s%s", song.GetTitle(), getColourCode(COLOUR_TEAL, COLOUR_DEFAULT), util.FormatSongLength(song.GetDuration()), getColourCode(COLOUR_DEFAULT, COLOUR_DEFAULT))
	return
}
