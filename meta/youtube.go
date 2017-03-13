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
		fmt.Errorf("%v", err)
		return
	}

	yt.service = service

	return
}

// GetMetaForItem - Get meta data for a youtube item
func (yt *YouTube) GetMetaForItem(source string) (meta *Meta, err error) {

	url, err := url.Parse(source)
	if err != nil {
		fmt.Errorf("Unable to parse source %v", err)
		return
	}

	identifier := url.Query().Get("v")
	if identifier == "" {
		return
	}

	call := yt.service.Videos.List("snippet,contentDetails").Id(identifier)
	response, err := call.Do()
	if err != nil {
		fmt.Errorf("youtube request failed %v", err)
		return
	}

	for _, item := range response.Items {
		if item.Id == identifier && item.Kind == "youtube#video" {

			d, convErr := duration.FromString(item.ContentDetails.Duration)
			if convErr != nil {
				fmt.Errorf("Unable to convert duration %v", convErr)
			}

			meta = &Meta{
				Identifier: identifier,
				Title:      item.Snippet.Title,
				Duration:   d.ToDuration(),
				Source:     source,
			}
		}
	}

	return
}
