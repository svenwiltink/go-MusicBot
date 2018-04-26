package mpv

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/dexterlb/mpvipc"
	"github.com/svenwiltink/go-musicbot/music"
	eventemitter "github.com/vansante/go-event-emitter"
)

const (
	mpvRetryAttempts  = 5
	mpvMaxLoadTimeout = time.Duration(time.Second * 20)
)

// MpvPlayer control MPV
type MpvPlayer struct {
	mpvMutex     sync.Mutex
	mpvProcess   *exec.Cmd
	mpvConn      *mpvipc.Connection
	mpvEvents    *eventemitter.Emitter
	mpvIsRunning bool

	mpvPath    string
	socketPath string
}

// NewPlayer creates an instance of MpvPlayer
func NewPlayer(mpvPath string, socketPath string) *MpvPlayer {
	return &MpvPlayer{
		mpvPath:    mpvPath,
		socketPath: socketPath,
	}
}

func (mpvPlayer *MpvPlayer) Start() error {
	mpvPlayer.mpvMutex.Lock()
	defer mpvPlayer.mpvMutex.Unlock()

	fi, err := os.Stat(mpvPlayer.socketPath)
	if err == nil && !fi.IsDir() {
		log.Printf("MpvControl.initMpv: Removing existing mpv input on: %s", mpvPlayer.socketPath)
		err = os.Remove(mpvPlayer.socketPath)
		if err != nil {
			log.Printf("MpvControl.initMpv: Error removing existing mpv input [%s] %v", mpvPlayer.socketPath, err)
			return err
		}
	}

	err = mpvPlayer.startProcess()
	if err != nil {
		log.Printf("MpvControl.initMpv: Error starting mpv [%s] %v", mpvPlayer.mpvPath, err)
		return err
	}

	attempts := 0
	for {
		// Give MPV a second or so to start and create the input socket
		time.Sleep(500 * time.Millisecond)

		log.Printf("MpvControl.initMpv: Attempting to open ipc connection to mpv [%s]", mpvPlayer.socketPath)
		mpvPlayer.mpvConn = mpvipc.NewConnection(mpvPlayer.socketPath)
		err = mpvPlayer.mpvConn.Open()
		if err != nil {
			if attempts >= mpvRetryAttempts {
				return fmt.Errorf("MpvControl.initMpv: Error opening ipc connection to mpv [%s] %v", mpvPlayer.socketPath, err)
			}
		} else {
			err = nil
			break
		}
		attempts++
	}

	log.Printf("MpvControl.initMpv: Connected to mpv ipc [%s]", mpvPlayer.socketPath)

	// Turn on all events.
	mpvPlayer.mpvConn.Call("enable_event", "all")

	mpvPlayer.mpvEvents.AddCapturer(func(eventName eventemitter.EventType, arguments ...interface{}) {
		var strArgs []string
		for _, arg := range arguments {
			strArgs = append(strArgs, fmt.Sprintf("%#v", arg))
		}
		log.Printf("MpvControl.mpvEvent [%s] %s", eventName, strings.Join(strArgs, " | "))
	})

	go func() {
		events, stopListening := mpvPlayer.mpvConn.NewEventListener()
		for event := range events {
			mpvPlayer.mpvEvents.EmitEvent("event", event)
		}
		stopListening <- struct{}{}
	}()

	return nil
}

func (mpvPlayer *MpvPlayer) startProcess() error {
	command := exec.Command(mpvPlayer.mpvPath, "--no-video", "--idle", "--input-ipc-server="+mpvPlayer.socketPath)
	mpvPlayer.mpvProcess = command

	err := command.Start()
	mpvPlayer.mpvIsRunning = err == nil

	if err != nil {
		return fmt.Errorf("MpvControl.startMpv: Error starting mpv [%s | %s] %v", mpvPlayer.mpvPath, mpvPlayer.socketPath, err)
	}

	go func() {
		err := command.Wait()

		mpvPlayer.mpvMutex.Lock()

		mpvPlayer.mpvIsRunning = false
		mpvPlayer.mpvProcess = nil

		mpvPlayer.mpvMutex.Unlock()

		log.Printf("MpvControl.startMpv: mpv has exited [%s | %s] %v", mpvPlayer.mpvPath, mpvPlayer.socketPath, err)
	}()

	return nil
}

func (mpvPlayer *MpvPlayer) CanPlay(song *music.Song) bool {
	return true
}

func (mpvPlayer *MpvPlayer) PlaySong(song *music.Song) error {
	mpvPlayer.mpvMutex.Lock()
	defer mpvPlayer.mpvMutex.Unlock()

	waitForLoad := make(chan bool)
	mpvPlayer.mpvEvents.ListenOnce("file-loaded", func(arguments ...interface{}) {
		waitForLoad <- true
	})

	// Start an event listener to wait for the file to load.
	_, err := mpvPlayer.mpvConn.Call("loadfile", song.Path, "replace")
	if err != nil {
		log.Printf("MpvControl.LoadFile: Error sending loadfile command [%s] %v", song.Path, err)
		return err
	}

	timeoutChan := time.NewTimer(mpvMaxLoadTimeout).C
	select {
	case <-waitForLoad:
		return nil
	case <-timeoutChan:
		log.Printf("MpvControl.LoadFile: Load file timeout, did not receive file-loaded event in %d", mpvMaxLoadTimeout)
		_, err = mpvPlayer.mpvConn.Call("stop")
		if err != nil {
			log.Printf("MpvControl.LoadFile: Error calling stop after timeout: %v", err)
			// TODO check if mpv is running
			return err
		}
		return fmt.Errorf("error loading file, mpv did not respond in time")
	}

	return nil
}

func (mpvPlayer *MpvPlayer) Play() error {
	panic("not implemented")
}

func (mpvPlayer *MpvPlayer) Pause() error {
	panic("not implemented")
}

func (mpvPlayer *MpvPlayer) Wait() {
	panic("not implemented")
}
