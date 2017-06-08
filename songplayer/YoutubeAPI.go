package songplayer

import (
	"errors"
	"fmt"
	"github.com/channelmeter/iso8601duration"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
	"net/http"
	"net/url"
)

const (
	YouTubeVideoURL    = "https://www.youtube.com/watch?v=%s"
	YouTubePlaylistURL = "https://www.youtube.com/watch?v=%s&list=%s"
)

type YouTubeAPI struct {
	apiKey  string
	service *youtube.Service
}

func NewYoutubeAPI(youtubeAPIKey string) (yt *YouTubeAPI) {
	yt = &YouTubeAPI{
		apiKey: youtubeAPIKey,
	}

	yt.init()
	return
}

// init - Initialize the youtube service
func (yt *YouTubeAPI) init() (err error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: yt.apiKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		logrus.Errorf("YoutubeAPI.init: Error creating client: %v", err)
		return
	}

	yt.service = service
	return
}

// GetPlayableForURL - Get meta data for a youtube url
func (yt *YouTubeAPI) GetPlayableForURL(source string) (playable Playable, err error) {
	ytURL, err := url.Parse(source)
	if err != nil {
		logrus.Errorf("YoutubeAPI.GetPlayableForURL: Unable to parse URL [%s] %v", source, err)
		return
	}

	identifier := ytURL.Query().Get("v")
	if identifier == "" {
		logrus.Errorf("YoutubeAPI.GetPlayableForURL: Empty identifier for: %s", source)
		err = fmt.Errorf("empty identifier for: %s", source)
		return
	}

	playable, err = yt.GetPlayableForIdentifier(identifier)
	if err != nil {
		logrus.Errorf("YoutubeAPI.GetPlayableForURL: Unable to get playable [%s] %v", identifier, err)
		return
	}
	return
}

func (yt *YouTubeAPI) GetPlayableForIdentifier(identifier string) (playable Playable, err error) {
	call := yt.service.Videos.List("snippet,contentDetails").Id(identifier)
	response, err := call.Do()
	if err != nil {
		logrus.Errorf("YoutubeAPI.GetPlayableForIdentifier: List request failed [%s] %v", identifier, err)
		return
	}

	for _, item := range response.Items {
		if item.Id == identifier && item.Kind == "youtube#video" {
			d, convErr := duration.FromString(item.ContentDetails.Duration)
			if convErr != nil {
				logrus.Errorf("YoutubeAPI.GetPlayableForIdentifier: Unable to convert duration [%s] %v", item.ContentDetails.Duration, convErr)
				return
			}
			if item.Snippet == nil {
				err = errors.New("snippet not found")
				return
			}

			imageURL := ""
			if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Medium != nil {
				imageURL = item.Snippet.Thumbnails.Medium.Url
			}

			playable = NewSong(item.Snippet.Title, d.ToDuration(), fmt.Sprintf(YouTubeVideoURL, identifier), imageURL)
			return
		}
	}
	err = fmt.Errorf("playable not found for: %s", identifier)
	return
}

func (yt *YouTubeAPI) GetPlayablesForPlaylistURL(source string) (items []Playable, err error) {
	plURL, err := url.Parse(source)
	if err != nil {
		logrus.Errorf("YoutubeAPI.GetPlayablesForPlaylistURL: Unable to parse URL [%s] %v", source, err)
		return
	}

	identifier := plURL.Query().Get("list")
	if identifier == "" {
		logrus.Errorf("YoutubeAPI.GetPlayablesForPlaylistURL: Empty list identifier for: %s", source)
		err = fmt.Errorf("empty list identifier for: %s", source)
		return
	}

	items, err = yt.GetPlayablesForPlaylistIdentifier(identifier, 100)
	return
}

func (yt *YouTubeAPI) GetPlayablesForPlaylistIdentifier(identifier string, limit int) (items []Playable, err error) {
	call := yt.service.PlaylistItems.List("snippet,contentDetails").MaxResults(int64(limit)).PlaylistId(identifier)
	response, err := call.Do()
	if err != nil {
		logrus.Errorf("YoutubeAPI.GetPlayablesForPlaylistIdentifier: Error listing [%s | %d] %v", identifier, limit, err)
		return
	}

	for _, item := range response.Items {
		if item.Kind != "youtube#playlistItem" {
			continue
		}
		item, err := yt.GetPlayableForIdentifier(item.ContentDetails.VideoId)
		if err == nil {
			items = append(items, item)
		}
	}
	return
}

func (yt *YouTubeAPI) Search(searchType SearchType, searchStr string, limit int) (items []PlayableSearchResult, err error) {
	searchTypeStr := "video"
	switch searchType {
	case SEARCH_TYPE_ALBUM:
		fallthrough
	case SEARCH_TYPE_PLAYLIST:
		searchTypeStr = "playlist"
	}

	call := yt.service.Search.List("id,snippet").
		Q(searchStr).
		Type(searchTypeStr).
		MaxResults(int64(limit))

	response, err := call.Do()
	if err != nil {
		logrus.Errorf("YoutubeAPI.Search: Error searching [%d | %s | %d] %v", searchType, searchStr, limit, err)
		return
	}

	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			var ply Playable
			ply, err = yt.GetPlayableForIdentifier(item.Id.VideoId)
			if err != nil {
				logrus.Errorf("YoutubeAPI.Search: Error getting playable [%s] %v", item.Id.VideoId, err)
				return
			}
			items = append(items, NewSongResult(SEARCH_TYPE_TRACK, ply.GetTitle(), ply.GetDuration(), ply.GetURL(), ply.GetImageURL()))
		case "youtube#playlist":
			var plys []Playable
			plys, err = yt.GetPlayablesForPlaylistIdentifier(item.Id.PlaylistId, 1)
			if err != nil || len(plys) < 1 {
				logrus.Errorf("YoutubeAPI.Search: Search request playlist lookup failed [%s]: %v", item.Id.PlaylistId, err)
				return
			}
			if item.Snippet == nil {
				err = errors.New("snippet not found")
				return
			}

			listURL := plys[0].GetURL() + "&list=" + item.Id.PlaylistId
			imageURL := ""
			if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Medium != nil {
				imageURL = item.Snippet.Thumbnails.Medium.Url
			}
			items = append(items, NewSongResult(SEARCH_TYPE_PLAYLIST, item.Snippet.Title, 0, listURL, imageURL))
		}
	}
	return
}
