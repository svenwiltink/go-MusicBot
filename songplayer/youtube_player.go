package songplayer

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/svenwiltink/go-musicbot/util"
	"regexp"
	"strings"
)

var youtubeURLRegex, _ = regexp.Compile(`^(https?://)?(www\.)?(youtube\.com|youtu\.?be)/.+$`)

type YoutubePlayer struct {
	*util.MpvControl

	ytAPI *YouTubeAPI
}

func NewYoutubePlayer(youtubeAPIKey string, mpvControl *util.MpvControl) (player *YoutubePlayer, err error) {
	if youtubeAPIKey == "" {
		err = errors.New("Youtube API key is empty")
		return
	}

	player = &YoutubePlayer{
		ytAPI: NewYoutubeAPI(youtubeAPIKey),
	}
	player.MpvControl = mpvControl
	return
}

func (p *YoutubePlayer) Name() (name string) {
	return "Youtube"
}

func (p *YoutubePlayer) CanPlay(url string) (canPlay bool) {
	return youtubeURLRegex.MatchString(url)
}

func (p *YoutubePlayer) GetSongs(url string) (songs []Playable, err error) {
	lowerURL := strings.ToLower(url)
	if strings.Contains(lowerURL, "player") || strings.Contains(lowerURL, "list=") {
		songs, err = p.ytAPI.GetPlayablesForPlaylistURL(url)
		// On error, fall back to single add
		if err == nil {
			return
		}
		logrus.Warnf("YoutubePlayer.GetSongs: Error getting playlist playables [%s] %v", url, err)
	}

	song, err := p.ytAPI.GetPlayableForURL(url)
	if err != nil {
		logrus.Errorf("YoutubePlayer.GetSongs: Error getting song playables [%s] %v", url, err)
		return
	}
	songs = append(songs, song)
	return
}

func (p *YoutubePlayer) Search(searchType SearchType, searchStr string, limit int) (results []PlayableSearchResult, err error) {
	results, err = p.ytAPI.Search(searchType, searchStr, limit)
	if err != nil {
		logrus.Errorf("YoutubePlayer.Search: Error searching songs [%d | %s | %d] %v", searchType, searchStr, limit, err)
		return
	}
	return
}

func (p *YoutubePlayer) Play(url string) (err error) {
	return p.LoadFile(url)
}
