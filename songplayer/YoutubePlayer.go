package songplayer

import (
	"fmt"
	"gitlab.transip.us/swiltink/go-MusicBot/meta"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

type YoutubePlayer struct {
	mpvBinPath   string
	mpvInputPath string

	mpvProcess   *exec.Cmd
	mpvIsRunning bool
	controlFile  *os.File
	ytService    *meta.YouTube
	mpvMutex     sync.Mutex
}

func NewYoutubePlayer(mpvBinPath, mpvInputPath string) (player *YoutubePlayer, err error) {
	player = &YoutubePlayer{
		mpvBinPath:   mpvBinPath,
		mpvInputPath: mpvInputPath,
		mpvIsRunning: false,
		ytService:    meta.NewYoutubeService(),
	}

	err = player.Init()
	return
}

func (p *YoutubePlayer) Init() (err error) {
	p.mpvMutex.Lock()
	defer p.mpvMutex.Unlock()

	syscall.Mknod(".mpv-input", syscall.S_IFIFO|0666, 0)
	file, err := os.OpenFile(".mpv-input", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		err = fmt.Errorf("[YoutubePlayer] Error opening control file: %v", err)
		return
	}

	p.controlFile = file
	return
}

func (p *YoutubePlayer) Name() (name string) {
	return "Youtube"
}

func (p *YoutubePlayer) CanPlay(url string) (canPlay bool) {
	return strings.Contains(strings.ToLower(url), "youtube")
}

func (p *YoutubePlayer) GetSongs(url string) (songs []Playable, err error) {
	lowerURL := strings.ToLower(url)
	if strings.Contains(lowerURL, "player") || strings.Contains(lowerURL, "list=") {
		var metaDatas []meta.Meta
		metaDatas, err = p.ytService.GetMetasForPlaylistURL(url)
		if err == nil {
			for _, metaData := range metaDatas {
				songs = append(songs, NewSong(metaData.Title, metaData.Duration, metaData.Source))
			}
			return
		}
		// On error, fall back to single add
	}

	metaData, err := p.ytService.GetMetaForURL(url)
	if err != nil {
		err = fmt.Errorf("[YoutubePlayer] Error getting meta data: %v", err)
		return
	}
	songs = append(songs, NewSong(metaData.Title, metaData.Duration, metaData.Source))
	return
}

func (p *YoutubePlayer) SearchSongs(searchStr string, limit int) (songs []Playable, err error) {
	metaDatas, err := p.ytService.SearchForMetas(searchStr, limit)
	if err != nil {
		err = fmt.Errorf("[YoutubePlayer] Error searching meta data: %v", err)
		return
	}

	for _, metaData := range metaDatas {
		songs = append(songs, NewSong(metaData.Title, metaData.Duration, metaData.Source))
	}
	return
}

func (p *YoutubePlayer) Play(url string) (err error) {
	p.mpvMutex.Lock()
	err = p.stop()
	if err != nil {
		return
	}

	command := exec.Command("mpv", "--no-video", "--input-file=.mpv-input", url)
	p.mpvProcess = command

	go func() {
		err = command.Start()
		p.mpvIsRunning = err == nil
		p.mpvMutex.Unlock()
		if err != nil {
			return
		}

		err = command.Wait()
		p.mpvIsRunning = false
	}()
	return
}

func (p *YoutubePlayer) Pause(pauseState bool) (err error) {
	p.mpvMutex.Lock()
	defer p.mpvMutex.Unlock()

	_, err = p.controlFile.WriteString("cycle pause\n")
	if err != nil {
		return
	}
	err = p.controlFile.Truncate(0)
	return
}

func (p *YoutubePlayer) Stop() (err error) {
	p.mpvMutex.Lock()
	defer p.mpvMutex.Unlock()

	return p.stop()
}

func (p *YoutubePlayer) stop() (err error) {
	if p.mpvIsRunning {
		err = p.mpvProcess.Process.Kill()
		p.mpvIsRunning = err == nil
	}
	return
}
