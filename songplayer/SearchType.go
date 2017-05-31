package songplayer

import (
	"strings"
)

type SearchType uint8

const (
	SEARCH_TYPE_TRACK    SearchType = 1
	SEARCH_TYPE_ALBUM               = 2
	SEARCH_TYPE_PLAYLIST            = 3
)

func GetSearchType(searchTypeStr string) (ok bool, searchType SearchType) {
	ok = true
	switch strings.ToLower(searchTypeStr) {
	case "track":
		searchType = SEARCH_TYPE_TRACK
	case "album":
		searchType = SEARCH_TYPE_ALBUM
	case "playlist":
		searchType = SEARCH_TYPE_PLAYLIST
	default:
		ok = false
	}
	return
}

func SearchTypeString(searchType SearchType) (str string) {
	switch searchType {
	case SEARCH_TYPE_TRACK:
		str = "track"
	case SEARCH_TYPE_ALBUM:
		str = "album"
	case SEARCH_TYPE_PLAYLIST:
		str = "playlist"
	}
	return
}
