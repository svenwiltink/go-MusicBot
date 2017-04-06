package player

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
	MpvProcess   *exec.Cmd
	mpvIsRunning bool
	ControlFile  *os.File
	ytService    *meta.YouTube
	mpvMutex     sync.Mutex
}

func NewYoutubePlayer() (player *YoutubePlayer, err error) {
	player = &YoutubePlayer{
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

	p.ControlFile = file
	return
}

func (p *YoutubePlayer) Name() (name string) {
	return "Youtube"
}

func (p *YoutubePlayer) CanPlay(url string) (canPlay bool) {
	return strings.Contains(strings.ToLower(url), "youtube")
}

func (p *YoutubePlayer) GetItems(url string) (items []ListItem, err error) {
	lowerURL := strings.ToLower(url)
	if strings.Contains(lowerURL, "playlist") || strings.Contains(lowerURL, "list=") {
		var metaDatas []meta.Meta
		metaDatas, err = p.ytService.GetMetasForPlaylistURL(url)
		if err == nil {
			for _, metaData := range metaDatas {
				items = append(items, *NewListItem(metaData.Title, metaData.Duration, metaData.Source))
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

	items = append(items, *NewListItem(metaData.Title, metaData.Duration, metaData.Source))
	return
}

func (p *YoutubePlayer) SearchItems(searchStr string, limit int) (items []ListItem, err error) {
	metaDatas, err := p.ytService.SearchForMetas(searchStr, limit)
	if err != nil {
		err = fmt.Errorf("[YoutubePlayer] Error searching meta data: %v", err)
		return
	}

	for _, metaData := range metaDatas {
		items = append(items, *NewListItem(metaData.Title, metaData.Duration, metaData.Source))
	}
	return
}

func (p *YoutubePlayer) Play(url string) (err error) {
	p.mpvMutex.Lock()
	if p.mpvIsRunning {
		fmt.Println("[YoutubePlayer] Killing Mpv")
		p.MpvProcess.Process.Kill()
		p.mpvIsRunning = false
		p.mpvMutex.Unlock()
		return
	}
	command := exec.Command("mpv", "--no-video", "--input-file=.mpv-input", url)
	p.MpvProcess = command

	go func() {
		command.Start()
		p.mpvIsRunning = true
		p.mpvMutex.Unlock()

		command.Wait()
		p.mpvIsRunning = false
	}()
	return
}

func (p *YoutubePlayer) Pause(pauseState bool) (err error) {
	p.mpvMutex.Lock()
	defer p.mpvMutex.Unlock()

	_, err = p.ControlFile.WriteString("cycle pause\n")
	if err != nil {
		return
	}
	err = p.ControlFile.Truncate(0)
	return
}

func (p *YoutubePlayer) Stop() (err error) {
	p.mpvMutex.Lock()
	defer p.mpvMutex.Unlock()

	if p.mpvIsRunning {
		fmt.Println("[YoutubePlayer] Killing mpv")
		p.MpvProcess.Process.Kill()
		p.mpvIsRunning = false
	}
	return
}
