package player

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type MpvPlayer struct {
	Queue        Queue
	CurrentSong  QueueItem
	Status       Status
	MpvProcess   *exec.Cmd
	mpvIsRunning bool
	ControlFile  *os.File
}

func NewMpvPlayer() *MpvPlayer {
	player := &MpvPlayer{
		Queue:        NewQueue(),
		Status:       STOPPED,
		mpvIsRunning: false,
	}

	player.init()

	return player
}

func (p *MpvPlayer) init() {

	syscall.Mknod(".mpv-input", syscall.S_IFIFO|0666, 0)
	file, err := os.OpenFile(".mpv-input", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		log.Fatal(err)
	}

	p.ControlFile = file
}

func (p *MpvPlayer) GetStatus() Status {
	return p.Status
}

func (p *MpvPlayer) Start() error {
	if p.Status != RUNNING {

		p.Status = RUNNING
		p.Next()

		return nil
	}

	return errors.New("Can't start a player that is already running")
}

func (p *MpvPlayer) Stop() {
	if p.Status == RUNNING {
		fmt.Println("Killing mpv")
		p.MpvProcess.Process.Kill()
	}
}

func (p *MpvPlayer) Pause() {
	p.ControlFile.WriteString("cycle pause\n")
	p.ControlFile.Truncate(0)
}

func (p *MpvPlayer) Next() {

	if p.mpvIsRunning {
		fmt.Println("Killing Mpv")
		p.MpvProcess.Process.Kill()
		return
	}

	if !p.Queue.HasNext() {
		p.Status = STOPPED
		p.CurrentSong = QueueItem{}
		return
	}

	nextSong, err := p.Queue.shift()
	if err != nil {
		fmt.Println(err)
		return
	}

	p.CurrentSong = nextSong
	url := nextSong.GetURL()
	command := exec.Command("mpv", "--no-video", "--input-file=.mpv-input", url)
	p.MpvProcess = command

	go func() {
		command.Start()
		p.mpvIsRunning = true
		command.Wait()
		p.mpvIsRunning = false
		p.Next()
	}()
}

func (p *MpvPlayer) AddSong(URL string, _ int64) {
	queueItem := NewQueueItem(URL)
	p.Queue.add(queueItem)

	// start the player if is not already running
	if p.Status == STOPPED {
		p.Start()
	}
}

func (p *MpvPlayer) GetCurrentSong() *QueueItem {
	return &p.CurrentSong
}

func (p *MpvPlayer) GetQueueItems() []QueueItem {
	return p.Queue.Items
}

func (p *MpvPlayer) FlushQueue() {
	p.Queue.Flush()
}

func (p *MpvPlayer) ShuffleQueue() {
	p.Queue.Shuffle()
}
