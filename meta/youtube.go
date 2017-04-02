package meta

import (
	"fmt"
	"github.com/channelmeter/iso8601duration"
	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
	"net/http"
	"net/url"
)

const APIKey = "AIzaSyAPEZOx4UgbBy6cEh_zZEfwYJ_3_bIWqfg"

const YTURL = "https://www.youtube.com/watch?v=%s"

type YouTube struct {
	service *youtube.Service
}

func NewYoutubeService() (y *YouTube) {
	y = &YouTube{}

	y.Initialize()
	return
}

// Initialize - Initialize the youtube service
func (yt *YouTube) Initialize() (err error) {
	client := &http.Client{
		Transport: &transport.APIKey{Key: APIKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		err = fmt.Errorf("[Youtube] Error creating meta client: %v", err)
		return
	}

	yt.service = service
	return
}

// GetMetaForURL - Get meta data for a youtube url
func (yt *YouTube) GetMetaForURL(source string) (meta *Meta, err error) {
	url, err := url.Parse(source)
	if err != nil {
		err = fmt.Errorf("[Youtube] ]Unable to parse source: %v", err)
		return
	}

	identifier := url.Query().Get("v")
	if identifier == "" {
		return
	}

	meta, err = yt.GetMetaForIdentifier(identifier)
	if err != nil {
		err = fmt.Errorf("[Youtube] Unable to get meta for source: %v", err)
		return
	}

	return
}

func (yt *YouTube) GetMetaForIdentifier(identifier string) (meta *Meta, err error) {
	call := yt.service.Videos.List("snippet,contentDetails").Id(identifier)
	response, err := call.Do()
	if err != nil {
		err = fmt.Errorf("[Youtube] Request failed: %v", err)
		return
	}

	meta = &Meta{}
	for _, item := range response.Items {
		if item.Id == identifier && item.Kind == "youtube#video" {
			d, convErr := duration.FromString(item.ContentDetails.Duration)
			if convErr != nil {
				err = fmt.Errorf("[Youtube] Unable to convert duration: %v", convErr)
				return
			}

			meta = &Meta{
				Identifier: identifier,
				Title:      item.Snippet.Title,
				Duration:   d.ToDuration(),
				Source:     fmt.Sprintf(YTURL, identifier),
			}
		}
	}
	return
}

func (yt *YouTube) GetMetasForPlaylistURL(source string) (items []Meta, err error) {
	url, err := url.Parse(source)
	if err != nil {
		err = fmt.Errorf("[Youtube] Unable to parse source: %v", err)
		return
	}

	identifier := url.Query().Get("list")
	if identifier == "" {
		return
	}

	call := yt.service.PlaylistItems.List("snippet,contentDetails").PlaylistId(identifier)
	response, err := call.Do()
	if err != nil {
		err = fmt.Errorf("[Youtube] Request failed: %v", err)
		return
	}

	for _, item := range response.Items {
		if item.Kind == "youtube#playlistItem" {
			item, err := yt.GetMetaForIdentifier(item.ContentDetails.VideoId)
			if err == nil {
				items = append(items, *item)
			}
		}
	}
	return
}

func (yt *YouTube) SearchForMetas(searchStr string, limit int) (items []Meta, err error) {
	call := yt.service.Search.List("id").
		Q(searchStr).
		Type("video").
		MaxResults(int64(limit))

	response, err := call.Do()
	if err != nil {
		err = fmt.Errorf("[Youtube] Search request failed: %v", err)
		return
	}

	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			meta, err := yt.GetMetaForIdentifier(item.Id.VideoId)
			if err == nil {
				items = append(items, *meta)
			}
		}
	}
	return
}
