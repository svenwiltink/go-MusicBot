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

func formatSong(song playlist.ItemInterface) (s string) {
	s = fmt.Sprintf("%s %s%s%s", song.GetTitle(), INVERSE_CHARACTER, util.FormatSongLength(song.GetDuration()), INVERSE_CHARACTER)
	return
}
