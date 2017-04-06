package api

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/vansante/go-event-emitter"
	"gitlab.transip.us/swiltink/go-MusicBot/playlist"
	"io/ioutil"
	"sync"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 128
)

type ControlWebsocket struct {
	ws        *websocket.Conn
	readOnly  bool
	hasClosed bool
	playlist  playlist.ListInterface
	capturer  *eventemitter.Capturer
	writeLock sync.Mutex
}

func NewControlWebsocket(ws *websocket.Conn, readOnly bool, playlist playlist.ListInterface) (cws *ControlWebsocket) {
	cws = &ControlWebsocket{
		ws:       ws,
		readOnly: readOnly,
		playlist: playlist,
	}
	return
}

func (cws *ControlWebsocket) Start() {
	cws.playlist.AddCapturer(cws.onEvent)

	go cws.socketWriter()
	cws.socketReader()
}

func (cws *ControlWebsocket) onEvent(event string, args ...interface{}) {
	for i := range args {
		itms, ok := args[i].([]playlist.ItemInterface)
		if ok {
			args[i] = getAPIItems(itms)
			continue
		}
		itm, ok := args[i].(playlist.ItemInterface)
		if ok {
			args[i] = getAPIItem(itm, itm.GetDuration())
			continue
		}
		dur, ok := args[i].(time.Duration)
		if ok {
			args[i] = int(dur.Seconds())
			continue
		}
	}
	evt := Event{
		Event:     event,
		Arguments: args,
	}
	cws.write(evt)
}

func (cws *ControlWebsocket) socketWriter() {
	ticker := time.NewTicker(pingPeriod)

	// Make sure we stop the ticker when done
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cws.writeLock.Lock()
			cws.ws.SetWriteDeadline(time.Now().Add(writeWait))
			err := cws.ws.WriteMessage(websocket.PingMessage, []byte{})
			cws.writeLock.Unlock()
			if err != nil {
				break
			}
		}
	}
	cws.closeConnection()
}

func (cws *ControlWebsocket) socketReader() {
	cws.ws.SetReadLimit(maxMessageSize)
	cws.ws.SetReadDeadline(time.Now().Add(pongWait))
	cws.ws.SetPongHandler(func(string) error {
		cws.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// If the application is not otherwise interested in messages from the peer,
	// then the application should start a goroutine to read and discard messages from the peer
	// (From the documentation)
	for {
		if tp, reader, err := cws.ws.NextReader(); err != nil {
			if cws.readOnly {
				// Ignore everything :)
				break
			}
			if tp != websocket.TextMessage {
				break
			}

			buf, err := ioutil.ReadAll(reader)
			if err != nil {
				cws.write("Error: " + err.Error())
				return
			}
			cmd := &Command{}
			err = json.Unmarshal(buf, cmd)
			if err != nil {
				cws.write("Error: " + err.Error())
				return
			}
			cws.executeCommand(cmd)
			break
		}
	}
	cws.closeConnection()
}

func (cws *ControlWebsocket) executeCommand(cmd *Command) {
	switch cmd.Command {
	case "add-items":
		if len(cmd.Arguments) < 1 {
			err := errors.New("Missing URL argument")
			cws.write(getCommandResponse(cmd, err))
			return
		}
		_, err := cws.playlist.AddItems(cmd.Arguments[0])
		cws.write(getCommandResponse(cmd, err))
	case "play":
		_, err := cws.playlist.Play()
		cws.write(getCommandResponse(cmd, err))
	case "pause":
		err := cws.playlist.Pause()
		cws.write(getCommandResponse(cmd, err))
	case "status":
		itm, remaining := cws.playlist.GetCurrentItem()
		resp := getCommandResponse(cmd, nil)
		resp.Status = &Status{
			Status:  cws.playlist.GetStatus(),
			Current: getAPIItem(itm, remaining),
			List:    getAPIItems(cws.playlist.GetItems()),
		}
		cws.write(resp)
	case "next":
		_, err := cws.playlist.Next()
		cws.write(getCommandResponse(cmd, err))
	case "stop":
		err := cws.playlist.Stop()
		cws.write(getCommandResponse(cmd, err))
	case "empty-list":
		cws.playlist.EmptyList()
		cws.write(getCommandResponse(cmd, nil))
	case "shuffle-list":
		cws.playlist.ShuffleList()
		cws.write(getCommandResponse(cmd, nil))
	}
}

func (cws *ControlWebsocket) closeConnection() {
	// Use the write lock to ensure this function is _really_ called once.
	cws.writeLock.Lock()
	defer cws.writeLock.Unlock()

	if cws.hasClosed {
		return
	}
	cws.hasClosed = true

	// Close websocket
	cws.ws.Close()

	// Remove our event capturer
	cws.playlist.RemoveCapturer(cws.capturer)
}

func (cws *ControlWebsocket) write(data interface{}) {
	// Applications are responsible for ensuring that no more than one goroutine calls the write methods
	// (From the docs)
	cws.writeLock.Lock()
	cws.ws.SetWriteDeadline(time.Now().Add(writeWait))
	cws.ws.WriteJSON(data)
	cws.writeLock.Unlock()
}
