package youtube

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	isoduration "github.com/channelmeter/iso8601duration"

	"github.com/svenwiltink/go-musicbot/pkg/music"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	youTubeVideoURL = "https://www.youtube.com/watch?v=%s&t=%d"
	MaxYoutubeItems = 500
)

var youtubeURLRegex = regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com|youtu\.?be)/.+$`)

type DataProvider struct {
	apiKey  string
	service *youtube.Service
}

func NewDataProvider(apiKey string) (*DataProvider, error) {
	instance := &DataProvider{
		apiKey: apiKey,
	}

	err := instance.initAPIClient()
	if err != nil {
		err = fmt.Errorf("could not start youtube api client: %v", err)
		return nil, err
	}

	return instance, nil
}

func (provider *DataProvider) initAPIClient() error {
	service, err := youtube.NewService(context.Background(), option.WithAPIKey(provider.apiKey))
	if err != nil {
		return fmt.Errorf("YoutubeAPI.init: Error creating client: %v", err)
	}

	provider.service = service

	return nil
}

func (provider *DataProvider) CanProvideData(song music.Song) bool {
	return youtubeURLRegex.MatchString(song.Path)
}

func (provider *DataProvider) ProvideData(song *music.Song) error {
	identifier, startTime, err := provider.getIdentifierAndStartTimeForSong(song)
	if err != nil {
		return err
	}

	return provider.provideDataForIdentifierAndStartTime(identifier, startTime, song)
}

func (provider *DataProvider) provideDataForIdentifierAndStartTime(identifier string, startTime int, song *music.Song) error {
	call := provider.service.Videos.List([]string{"snippet", "contentDetails"}).Id(identifier)
	response, err := call.Do()
	if err != nil {
		return fmt.Errorf("could not get data for url: %v", err)
	}

	for _, item := range response.Items {
		if item.Id == identifier && item.Kind == "youtube#video" {
			if item.Snippet == nil {
				return errors.New("snippet not found")
			}

			song.Name = item.Snippet.Title
			song.Artist = item.Snippet.ChannelTitle

			duration, err := isoduration.FromString(item.ContentDetails.Duration)

			if err != nil {
				return err
			}

			song.Duration = duration.ToDuration()

			song.Path = fmt.Sprintf(youTubeVideoURL, identifier, startTime)
			return nil
		}
	}

	return fmt.Errorf("playable not found for: %s", identifier)
}

func (provider *DataProvider) getIdentifierAndStartTimeForSong(song *music.Song) (string, int, error) {
	ytURL, err := url.Parse(song.Path)
	if err != nil {
		return "", 0, fmt.Errorf("YoutubeAPI.GetPlayableForURL: Unable to parse URL [%s] %v", song.Path, err)
	}

	identifier := ytURL.Query().Get("v")
	if identifier == "" {
		// Assume format like: https://youtu.be/n1dpZy5Jx4o, in which the path is the identifier
		identifier = strings.TrimLeft(ytURL.Path, "/")

		if strings.ToLower(identifier) == "watch" || identifier == "" {
			return "", 0, fmt.Errorf("empty identifier for: %s", song.Path)
		}
	}
	startTimeString := ytURL.Query().Get("t")
	startTimeString = strings.TrimRight(startTimeString, "s")
	startTime, startTimeErr := strconv.Atoi(startTimeString)
	if startTimeErr != nil {
		startTime = 0
	}

	return identifier, startTime, nil
}

func (provider *DataProvider) Search(searchString string) ([]music.Song, error) {
	searchTypeStr := "video"

	call := provider.service.Search.List([]string{"id", "snippet"}).
		Q(searchString).
		Type(searchTypeStr).
		MaxResults(int64(5))

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("YoutubeApi: error searching %s: %v", searchString, err)
	}

	songs := make([]music.Song, 0)

	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			song := &music.Song{}
			err = provider.provideDataForIdentifierAndStartTime(item.Id.VideoId, 0, song)
			if err != nil {
				return nil, fmt.Errorf("error finding data: %v", err)
			}

			songs = append(songs, *song)
		}
	}
	return songs, nil
}
