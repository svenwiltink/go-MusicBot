package MusicPlayer

import (
	"fmt"
	"os/exec"
	"os"
	"log"
	"errors"
)

type PlayerStatus int

const (
	RUNNING PlayerStatus = 1 + iota
	STOPPED
)

type MusicPlayer struct {
	Queue Queue
	CurrentSong QueueItem
	Status PlayerStatus
	MpvProcess *exec.Cmd
	ControlFile *os.File
}

func NewMusicPlayer() *MusicPlayer {
	player := &MusicPlayer{
		Queue: NewQueue(),
		Status: STOPPED,
	}

	player.init()

	return player
}

func (p *MusicPlayer) init() {
	file, err := os.OpenFile(".mpv-input", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}

	p.ControlFile = file
}

func (p *MusicPlayer) GetStatus() PlayerStatus {
	return p.Status
}

func (p *MusicPlayer) Start() error {
	if p.Status != RUNNING {
		p.Status = RUNNING
		p.Next()
		return nil
	} else {
		return errors.New("Can't start a player that is already running")
	}
}

func (p *MusicPlayer) Stop() {
	if p.Status == RUNNING {
		fmt.Println("Killing mpv")
		p.MpvProcess.Process.Kill()
	}
}

func (p *MusicPlayer) Pause() {
	p.ControlFile.WriteString("cycle pause\n")
	p.ControlFile.Truncate(0)
}

func (p *MusicPlayer) Next() {

	if p.MpvProcess != nil && !p.MpvProcess.ProcessState.Exited() {
		fmt.Println("Killing Mpv")
		p.MpvProcess.Process.Kill()
	}

	if !p.Queue.HasNext() {
		p.Status = STOPPED
		return
	}

	nextSong, err := p.Queue.shift();
	if err != nil {
		fmt.Println(err)
		return
	}

	url := nextSong.GetUrl();
	//command := exec.Command("mpv", "--no-video", "--input-file=.mpv-input", url)
	command := exec.Command("mpv", "--input-file=.mpv-input", url)
	p.MpvProcess = command
	go func(player *MusicPlayer, command *exec.Cmd) {
		command.Start()
		command.Wait()
		player.Next()
	}(p, command)
}

func (p *MusicPlayer) AddSong(Url string) {
	queueItem := NewQueueItem(Url)
	p.Queue.add(queueItem)

	// start the player if is not already running
	if (p.Status == STOPPED) {
		p.Start()
	}
}