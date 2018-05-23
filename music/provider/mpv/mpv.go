package mpv

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/DexterLB/mpvipc"
	"github.com/svenwiltink/go-musicbot/music"
	eventemitter "github.com/vansante/go-event-emitter"
	"context"
)

const (
	mpvRetryAttempts  = 5
	mpvMaxLoadTimeout = time.Duration(time.Second * 5)
)

// MPV events
const (
	EventFileLoaded eventemitter.EventType = "file-loaded"
	EventFileEnded  eventemitter.EventType = "end-file"
)

// Player control MPV
type Player struct {
	mutex        sync.Mutex
	process      *exec.Cmd
	connection   *mpvipc.Connection
	eventEmitter *eventemitter.Emitter
	isRunning    bool

	mpvPath    string
	socketPath string
}
// NewPlayer creates an instance of MpvPlayer
func NewPlayer(mpvPath string, socketPath string) *Player {
	return &Player{
		mpvPath:      mpvPath,
		socketPath:   socketPath,
		isRunning:    false,
		eventEmitter: eventemitter.NewEmitter(true),
	}
}

func (player *Player) SetVolume(percentage int) {
	player.mutex.Lock()
	defer player.mutex.Unlock()

	player.connection.Call("set_property", "volume", percentage)
}

// Start the mpv player
func (player *Player) Start() error {
	player.mutex.Lock()
	defer player.mutex.Unlock()

	err := player.removeExistingFile()
	if err != nil {
		log.Printf("Mpv: Error starting mpv [%s] %v", player.mpvPath, err)
		return err
	}

	err = player.startProcess()
	if err != nil {
		log.Printf("Mpv: Error starting mpv [%s] %v", player.mpvPath, err)
		return err
	}

	err = player.waitForMpv()
	if err != nil {
		return err
	}

	player.startEventListeners()

	return nil
}

func (player *Player) removeExistingFile() error {
	fi, err := os.Stat(player.socketPath)
	if err == nil && !fi.IsDir() {
		log.Printf("Mpv: Removing existing mpv input on: %s", player.socketPath)
		err = os.Remove(player.socketPath)
		if err != nil {
			log.Printf("Mpv: Error removing existing mpv input [%s] %v", player.socketPath, err)
			return err
		}
	}

	return nil
}

func (player *Player) waitForMpv() error {
	attempts := 0
	for {
		// Give MPV a second or so to start and create the input socket
		time.Sleep(500 * time.Millisecond)

		log.Printf("Mpv: Attempting to open ipc connection to mpv [%s]", player.socketPath)
		player.connection = mpvipc.NewConnection(player.socketPath)
		err := player.connection.Open()

		if err == nil {
			return nil
		}

		if attempts >= mpvRetryAttempts {
			return fmt.Errorf("Mpv: Error opening ipc connection to mpv [%s] %v", player.socketPath, err)
		}

		attempts++
	}
}

func (player *Player) startEventListeners() {
	player.eventEmitter.AddCapturer(func(eventName eventemitter.EventType, arguments ...interface{}) {
		if len(eventName) == 0 {
			return
		}

		var strArgs []string
		for _, arg := range arguments {
			strArgs = append(strArgs, fmt.Sprintf("%#v", arg))
		}
		log.Printf("MpvControl.mpvEvent [%s] %s", eventName, strings.Join(strArgs, " | "))
	})

	go func() {
		events, stopListening := player.connection.NewEventListener()
		for event := range events {
			player.eventEmitter.EmitEvent(eventemitter.EventType(event.Name), event)
		}
		stopListening <- struct{}{}
	}()
}

// Skip the current song
func (player *Player) Skip() error {
	player.mutex.Lock()
	defer player.mutex.Unlock()

	_, err := player.connection.Call("stop")

	return err
}

func (player *Player) startProcess() error {
	command := exec.Command(player.mpvPath, "--no-video", "--idle", "--input-ipc-server="+player.socketPath)
	player.process = command

	err := command.Start()
	player.isRunning = err == nil

	if err != nil {
		return fmt.Errorf("MpvControl.startMpv: Error starting mpv [%s | %s] %v", player.mpvPath, player.socketPath, err)
	}

	go func() {
		err := command.Wait()

		player.mutex.Lock()

		player.isRunning = false
		player.process = nil

		player.mutex.Unlock()

		log.Printf("MpvControl.startMpv: mpv has exited [%s | %s] %v", player.mpvPath, player.socketPath, err)
	}()

	return nil
}

func (player *Player) CanPlay(song *music.Song) bool {
	return true
}

func (player *Player) PlaySong(song *music.Song) error {
	player.mutex.Lock()
	defer player.mutex.Unlock()

	waitForLoad := make(chan bool)
	player.eventEmitter.ListenOnce(EventFileLoaded, func(arguments ...interface{}) {
		waitForLoad <- true
	})

	// Start an event listener to wait for the file to load.
	_, err := player.connection.Call("loadfile", song.Path, "replace")
	if err != nil {
		log.Printf("MpvControl.LoadFile: Error sending loadfile command [%s] %v", song.Path, err)
		return err
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), mpvMaxLoadTimeout)
	defer cancel()

	select {
	case <-waitForLoad:
		return nil
	case <-timeoutCtx.Done():
		log.Printf("MpvControl.LoadFile: Load file timeout, did not receive file-loaded event in %d", mpvMaxLoadTimeout)
		_, err = player.connection.Call("stop")
		if err != nil {
			log.Printf("MpvControl.LoadFile: Error calling stop after timeout: %v", err)
			return err
		}
		return fmt.Errorf("error loading file, mpv did not respond in time")
	}

	return nil
}

func (player *Player) Play() error {
	panic("not implemented")
}

func (player *Player) Pause() error {
	panic("not implemented")
}

func (player *Player) Stop() {
	if player.isRunning {
		player.process.Process.Kill()
	}

	os.Remove(player.socketPath)
}

func (player *Player) Wait() {

	done := make(chan bool)
	player.eventEmitter.ListenOnce(EventFileEnded, func(arguments ...interface{}) {
		done <- true
	})

	<-done
}
