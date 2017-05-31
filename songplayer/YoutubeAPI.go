package songplayer

import (
	"errors"
	"fmt"
	"github.com/channelmeter/iso8601duration"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
	"net/http"
	"net/url"
)

const APIKey = "AIzaSyAPEZOx4UgbBy6cEh_zZEfwYJ_3_bIWqfg"

const (
	YouTubeVideoURL    = "https://www.youtube.com/watch?v=%s"
	YouTubePlaylistURL = "https://www.youtube.com/watch?v=%s&list=%s"
)

type YouTubeAPI struct {
	service *youtube.Service
}

func NewYoutubeAPI() (y *YouTubeAPI) {
	y = &YouTubeAPI{}

	y.Initialize()
	return
}

// Initialize - Initialize the youtube service
func (yt *YouTubeAPI) Initialize() (err error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: APIKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		err = fmt.Errorf("[YoutubeMeta] Error creating meta client: %v", err)
		return
	}

	yt.service = service
	return
}

// GetPlayableForURL - Get meta data for a youtube url
func (yt *YouTubeAPI) GetPlayableForURL(source string) (playable Playable, err error) {
	ytURL, err := url.Parse(source)
	if err != nil {
		err = fmt.Errorf("[YoutubeMeta] Unable to parse source: %v", err)
		return
	}

	identifier := ytURL.Query().Get("v")
	if identifier == "" {
		err = fmt.Errorf("[YoutubeMeta] Empty identifier for: %s", source)
		return
	}

	playable, err = yt.GetPlayableForIdentifier(identifier)
	if err != nil {
		err = fmt.Errorf("[YoutubeMeta] Unable to get meta for source: %v", err)
		return
	}
	return
}

func (yt *YouTubeAPI) GetPlayableForIdentifier(identifier string) (playable Playable, err error) {
	call := yt.service.Videos.List("snippet,contentDetails").Id(identifier)
	response, err := call.Do()
	if err != nil {
		err = fmt.Errorf("[YoutubeMeta] Request failed: %v", err)
		return
	}

	for _, item := range response.Items {
		if item.Id == identifier && item.Kind == "youtube#video" {
			d, convErr := duration.FromString(item.ContentDetails.Duration)
			if convErr != nil {
				err = fmt.Errorf("[YoutubeMeta] Unable to convert duration: %v", convErr)
				return
			}
			if item.Snippet == nil {
				err = errors.New("[YoutubeMeta] Snippet not found")
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
	err = fmt.Errorf("[YoutubeMeta] Meta not found for: %s", identifier)
	return
}

func (yt *YouTubeAPI) GetPlayablesForPlaylistURL(source string) (items []Playable, err error) {
	plURL, err := url.Parse(source)
	if err != nil {
		err = fmt.Errorf("[YoutubeMeta] Unable to parse source: %v", err)
		return
	}

	identifier := plURL.Query().Get("list")
	if identifier == "" {
		return
	}

	items, err = yt.GetPlayablesForPlaylistIdentifier(identifier, 100)
	return
}

func (yt *YouTubeAPI) GetPlayablesForPlaylistIdentifier(identifier string, limit int) (items []Playable, err error) {
	call := yt.service.PlaylistItems.List("snippet,contentDetails").MaxResults(int64(limit)).PlaylistId(identifier)
	response, err := call.Do()
	if err != nil {
		err = fmt.Errorf("[YoutubeMeta] Request failed: %v", err)
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

	call := yt.service.Search.List("id").
		Q(searchStr).
		Type(searchTypeStr).
		MaxResults(int64(limit))

	response, err := call.Do()
	if err != nil {
		err = fmt.Errorf("[YoutubeMeta] Search request failed: %v", err)
		return
	}

	fmt.Printf("Results: %v", response)
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			var ply Playable
			ply, err = yt.GetPlayableForIdentifier(item.Id.VideoId)
			if err != nil {
				err = fmt.Errorf("[YoutubeMeta] Search request video lookup failed [%s]: %v", item.Id.VideoId, err)
				return
			}
			items = append(items, NewSongResult(SEARCH_TYPE_TRACK, ply.GetTitle(), ply.GetDuration(), ply.GetURL(), ply.GetImageURL()))
		case "youtube#playlist":
			var plys []Playable
			plys, err = yt.GetPlayablesForPlaylistIdentifier(item.Id.PlaylistId, 1)
			if err != nil || len(plys) < 1 {
				err = fmt.Errorf("[YoutubeMeta] Search request playlist lookup failed [%s]: %v", item.Id.PlaylistId, err)
				return
			}
			if item.Snippet == nil {
				err = errors.New("[YoutubeMeta] Snippet not found")
				return
			}

			listURL := fmt.Sprintf(YouTubePlaylistURL, plys[0].GetURL(), item.Id.PlaylistId)
			imageURL := ""
			if item.Snippet.Thumbnails != nil && item.Snippet.Thumbnails.Medium != nil {
				imageURL = item.Snippet.Thumbnails.Medium.Url
			}
			items = append(items, NewSongResult(SEARCH_TYPE_PLAYLIST, item.Snippet.Title, 0, listURL, imageURL))
		}
	}
	return
}
