package api

import (
	"encoding/json"
	"errors"
	"github.com/SvenWiltink/go-MusicBot/player"
	"github.com/SvenWiltink/go-MusicBot/songplayer"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/vansante/go-event-emitter"
	"io/ioutil"
	"strconv"
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
	player    player.MusicPlayer
	capturer  *eventemitter.Capturer
	writeLock sync.Mutex
	user      string
}

func NewControlWebsocket(ws *websocket.Conn, readOnly bool, player player.MusicPlayer, user string) (cws *ControlWebsocket) {
	cws = &ControlWebsocket{
		ws:       ws,
		readOnly: readOnly,
		player:   player,
		user:     user,
	}
	return
}

func (cws *ControlWebsocket) Start() {
	cws.capturer = cws.player.AddCapturer(cws.onEvent)

	go cws.socketWriter()
	cws.socketReader()
}

func (cws *ControlWebsocket) onEvent(event string, args ...interface{}) {
	apiArgs := make([]interface{}, len(args))
	// Loop through the arguments and if its a known type, translate it to the appropiate API type.
	for i := range args {
		apiArgs[i] = args[i]
		songs, ok := apiArgs[i].([]songplayer.Playable)
		if ok {
			apiArgs[i] = getAPISongs(songs)
			continue
		}
		song, ok := apiArgs[i].(songplayer.Playable)
		if ok {
			apiArgs[i] = getAPISong(song, song.GetDuration())
			continue
		}
		plyr, ok := apiArgs[i].(songplayer.SongPlayer)
		if ok {
			apiArgs[i] = plyr.Name()
			continue
		}
		dur, ok := apiArgs[i].(time.Duration)
		if ok {
			apiArgs[i] = int(dur.Seconds())
			continue
		}
	}
	evt := Event{
		Event:     event,
		Arguments: apiArgs,
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
				logrus.Errorf("ControlWebSocket.socketReader: Error reading: %v", err)
				cws.write("Error: " + err.Error())
				return
			}
			cmd := &Command{}
			err = json.Unmarshal(buf, cmd)
			if err != nil {
				logrus.Errorf("ControlWebSocket.socketReader: Error unmarshalling: %v", err)
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
	case "add":
		if len(cmd.Arguments) < 1 {
			err := errors.New("Missing URL argument")
			cws.write(getCommandResponse(cmd, err))
			return
		}
		_, err := cws.player.Add(cmd.Arguments[0], cws.user)
		cws.write(getCommandResponse(cmd, err))
	case "open":
		if len(cmd.Arguments) < 1 {
			err := errors.New("Missing URL argument")
			cws.write(getCommandResponse(cmd, err))
			return
		}
		_, err := cws.player.Insert(cmd.Arguments[0], 0, cws.user)
		cws.write(getCommandResponse(cmd, err))
	case "play":
		_, err := cws.player.Play()
		cws.write(getCommandResponse(cmd, err))
	case "pause":
		err := cws.player.Pause()
		cws.write(getCommandResponse(cmd, err))
	case "status":
		song, remaining := cws.player.GetCurrent()
		resp := getCommandResponse(cmd, nil)
		resp.Status = &Status{
			Status:  cws.player.GetStatus(),
			Current: getAPISong(song, remaining),
			List:    getAPISongs(cws.player.GetQueue()),
		}
		cws.write(resp)
	case "next":
		_, err := cws.player.Next()
		cws.write(getCommandResponse(cmd, err))
	case "previous":
		_, err := cws.player.Previous()
		cws.write(getCommandResponse(cmd, err))
	case "jump":
		if len(cmd.Arguments) < 1 {
			err := errors.New("Missing deltaIndex argument")
			cws.write(getCommandResponse(cmd, err))
			return
		}
		index, err := strconv.ParseInt(cmd.Arguments[0], 10, 32)
		if err != nil {
			cws.write(getCommandResponse(cmd, err))
			return
		}
		_, err = cws.player.Jump(int(index))
		cws.write(getCommandResponse(cmd, err))
	case "stop":
		err := cws.player.Stop()
		cws.write(getCommandResponse(cmd, err))
	case "empty-list":
		cws.player.EmptyQueue()
		cws.write(getCommandResponse(cmd, nil))
	case "shuffle-list":
		cws.player.ShuffleQueue()
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
	cws.player.RemoveCapturer(cws.capturer)
}

func (cws *ControlWebsocket) write(data interface{}) {
	// Applications are responsible for ensuring that no more than one goroutine calls the write methods
	// (From the docs)
	cws.writeLock.Lock()
	cws.ws.SetWriteDeadline(time.Now().Add(writeWait))
	cws.ws.WriteJSON(data)
	cws.writeLock.Unlock()
}
