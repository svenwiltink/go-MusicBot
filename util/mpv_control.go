package util

import (
	"fmt"
	"github.com/dexterlb/mpvipc"
	"github.com/sirupsen/logrus"
	"github.com/vansante/go-event-emitter"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	MPV_INIT_RETRY_ATTEMPTS = 5
	MAX_MPV_LOAD_WAIT       = time.Duration(time.Second * 20)
)

type MpvControl struct {
	mpvBinPath   string
	mpvInputPath string

	mpvProcess   *exec.Cmd
	mpvIsRunning bool
	mpvConn      *mpvipc.Connection
	mpvMutex     sync.Mutex
	mpvEvents    *eventemitter.Emitter
}

func NewMpvControl(mpvBinPath, mpvInputPath string) (control *MpvControl, err error) {
	if mpvBinPath == "" {
		mpvBinPath = "mpv"
	}
	if mpvInputPath == "" {
		mpvInputPath = ".mpv-input"
	}
	control = &MpvControl{
		mpvBinPath:   mpvBinPath,
		mpvInputPath: mpvInputPath,
		mpvIsRunning: false,
		mpvEvents:    eventemitter.NewEmitter(),
	}

	err = control.initMpv()
	return
}

func (c *MpvControl) initMpv() (err error) {
	c.mpvMutex.Lock()
	defer c.mpvMutex.Unlock()

	fi, err := os.Stat(c.mpvInputPath)
	if err == nil && !fi.IsDir() {
		logrus.Infof("MpvControl.initMpv: Removing existing mpv input on: %s", c.mpvInputPath)
		err = os.Remove(c.mpvInputPath)
		if err != nil {
			logrus.Errorf("MpvControl.initMpv: Error removing existing mpv input [%s] %v", c.mpvInputPath, err)
			return
		}
	}

	err = c.startMpv()
	if err != nil {
		logrus.Errorf("MpvControl.initMpv: Error starting mpv [%s] %v", c.mpvBinPath, err)
		return
	}

	attempts := 0
	for {
		// Give MPV a second or so to start and create the input socket
		time.Sleep(500 * time.Millisecond)

		logrus.Infof("MpvControl.initMpv: Attempting to open ipc connection to mpv [%s]", c.mpvInputPath)
		c.mpvConn = mpvipc.NewConnection(c.mpvInputPath)
		err = c.mpvConn.Open()
		if err != nil {
			if attempts >= MPV_INIT_RETRY_ATTEMPTS {
				logrus.Errorf("MpvControl.initMpv: Error opening ipc connection to mpv [%s] %v", c.mpvInputPath, err)
				return
			}
		} else {
			err = nil
			break
		}
		attempts++
	}

	logrus.Infof("MpvControl.initMpv: Connected to mpv ipc [%s]", c.mpvInputPath)

	// Turn on all events.
	c.mpvConn.Call("enable_event", "all")

	go func() {
		events, stopListening := c.mpvConn.NewEventListener()
		for event := range events {
			c.mpvEvents.EmitEvent(event.Name, event)
		}
		stopListening <- struct{}{}
	}()

	return
}

func (c *MpvControl) startMpv() (err error) {
	logrus.Infof("MpvControl.startMpv: Starting MPV %s with control %s in idle mode", c.mpvBinPath, c.mpvInputPath)

	command := exec.Command(c.mpvBinPath, "--no-video", "--idle", "--input-ipc-server="+c.mpvInputPath)
	c.mpvProcess = command

	err = command.Start()
	c.mpvIsRunning = err == nil

	if err != nil {
		logrus.Errorf("MpvControl.startMpv: Error starting mpv [%s | %s] %v", c.mpvBinPath, c.mpvInputPath, err)
		return
	}

	go func() {
		err := command.Wait()

		c.mpvMutex.Lock()
		c.mpvIsRunning = false
		c.mpvProcess = nil
		c.mpvMutex.Unlock()

		logrus.Infof("MpvControl.startMpv: mpv has exited [%s | %s] %v", c.mpvBinPath, c.mpvInputPath, err)
	}()
	return
}

func (c *MpvControl) checkRunning() (err error) {
	if c.mpvIsRunning && c.mpvProcess != nil {
		return
	}
	logrus.Warn("MpvControl.checkRunning: mpv is not running, restarting")
	err = c.initMpv()
	if err != nil {
		logrus.Errorf("MpvControl.checkRunning: Error restarting mpv: %v", err)
		return
	}
	return
}

func (c *MpvControl) LoadFile(uri string) (err error) {
	c.mpvMutex.Lock()
	defer c.mpvMutex.Unlock()

	err = c.checkRunning()
	if err != nil {
		logrus.Errorf("MpvControl.LoadFile: Running check error: %v", err)
		return
	}

	err = c.stop()
	if err != nil {
		return
	}

	waitForLoad := make(chan bool)
	c.mpvEvents.ListenOnce("file-loaded", func(arguments ...interface{}) {
		waitForLoad <- true
	})

	// Start an event listener to wait for the file to load.
	_, err = c.mpvConn.Call("loadfile", uri, "replace")
	if err != nil {
		logrus.Errorf("MpvControl.LoadFile: Error sending loadfile command [%s] %v", uri, err)
		return
	}

	timeoutChan := time.NewTimer(MAX_MPV_LOAD_WAIT).C
	select {
	case <-waitForLoad:
		return
	case <-timeoutChan:
		logrus.Warnf("MpvControl.LoadFile: Load file timeout, did not receive file-loaded event in %d", MAX_MPV_LOAD_WAIT)
		_, err = c.mpvConn.Call("stop")
		if err != nil {
			logrus.Errorf("MpvControl.LoadFile: Error calling stop after timeout: %v", err)
			c.checkRunning()
			return
		}
		err = fmt.Errorf("error loading file, mpv did not respond in time")
		return
	}
	return
}

func (c *MpvControl) Seek(positionSeconds int) (err error) {
	c.mpvMutex.Lock()
	defer c.mpvMutex.Unlock()

	err = c.checkRunning()
	if err != nil {
		logrus.Errorf("MpvControl.Seek: Running check error: %v", err)
		return
	}

	err = c.mpvConn.Set("time-pos", positionSeconds)
	if err != nil {
		logrus.Errorf("MpvControl.Seek: Error sending time-pos command [%d] %v", positionSeconds, err)
		return
	}
	return
}

func (c *MpvControl) Pause(pauseState bool) (err error) {
	c.mpvMutex.Lock()
	defer c.mpvMutex.Unlock()

	err = c.checkRunning()
	if err != nil {
		logrus.Errorf("MpvControl.Pause: Running check error: %v", err)
		return
	}

	err = c.mpvConn.Set("pause", pauseState)
	if err != nil {
		logrus.Errorf("MpvControl.Pause: Error sending pause command [%v] %v", pauseState, err)
		return
	}
	return
}

func (c *MpvControl) Stop() (err error) {
	c.mpvMutex.Lock()
	defer c.mpvMutex.Unlock()

	return c.stop()
}

func (c *MpvControl) stop() (err error) {
	err = c.checkRunning()
	if err != nil {
		logrus.Errorf("MpvControl.stop: Running check error: %v", err)
		return
	}

	_, err = c.mpvConn.Call("stop")
	if err != nil {
		logrus.Errorf("MpvControl.stop: Error sending stop command: %v", err)
		return
	}
	return
}
