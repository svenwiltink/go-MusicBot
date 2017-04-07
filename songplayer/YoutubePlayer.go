package songplayer

import (
	"fmt"
	"gitlab.transip.us/swiltink/go-MusicBot/meta"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
)

var youtubeURLRegex, _ = regexp.Compile(`^(https?\:\/\/)?(www\.)?(youtube\.com|youtu\.?be)\/.+$`)

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

	fmt.Printf("[YoutubePlayer] Creating MPV control node on %s\n", p.mpvInputPath)
	syscall.Mknod(p.mpvInputPath, syscall.S_IFIFO|0666, 0)
	file, err := os.OpenFile(p.mpvInputPath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
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
	return youtubeURLRegex.MatchString(url)
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

	fmt.Printf("[YoutubePlayer] Starting MPV %s with control %s and url %s\n", p.mpvBinPath, p.mpvInputPath, url)
	command := exec.Command(p.mpvBinPath, "--no-video", "--input-file="+p.mpvInputPath, url)
	p.mpvProcess = command

	err = command.Start()
	p.mpvIsRunning = err == nil
	p.mpvMutex.Unlock()
	if err != nil {
		return
	}

	go func() {
		err := command.Wait()
		if err != nil {
			fmt.Printf("[YoutubePlayer] Error while waiting for mpv: %v\n", err)
		}
		p.mpvIsRunning = false
	}()
	return
}

func (p *YoutubePlayer) Pause(pauseState bool) (err error) {
	p.mpvMutex.Lock()
	defer p.mpvMutex.Unlock()

	fmt.Printf("[YoutubePlayer] Sending MPV control %s pause command\n", p.mpvInputPath)
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
