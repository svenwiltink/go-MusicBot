package mattermost

import (
	"fmt"
	"github.com/gorilla/websocket"
	mattermost "github.com/mattermost/mattermost-server/v5/model"
	"github.com/svenwiltink/go-musicbot/pkg/bot"
	"log"
	"strings"
	"time"
)

type MessageProvider struct {
	Config         *bot.Config
	MessageChannel chan bot.Message

	team    *mattermost.Team
	channel *mattermost.Channel

	client          *mattermost.Client4
	websocketClient *mattermost.WebSocketClient
}

func (provider *MessageProvider) Start() error {

	protocol := "http://"
	if provider.Config.Mattermost.Ssl {
		protocol = "https://"
	}

	provider.client = mattermost.NewAPIv4Client(protocol + provider.Config.Mattermost.Server)
	provider.client.SetOAuthToken(provider.Config.Mattermost.PrivateAccessToken)

	team, response := provider.client.GetTeamByName(provider.Config.Mattermost.Teamname, "")
	if response.Error != nil {
		return fmt.Errorf("unable to get team by name %s: %+v", provider.Config.Mattermost.Teamname, response.Error)
	}

	provider.team = team

	channel, response := provider.client.GetChannelByName(provider.Config.Mattermost.Channel, team.Id, "")
	if response.Error != nil {
		return fmt.Errorf("unable to get channel by name %s: %+v", provider.Config.Mattermost.Teamname, response.Error)
	}

	provider.channel = channel

	err := provider.connect()
	if err != nil {
		return err
	}

	go provider.pingLoop()
	go provider.startReadLoop()

	return nil
}

func (provider *MessageProvider) SendReplyToMessage(message bot.Message, reply string) error {
	post := &mattermost.Post{
		ChannelId: message.Target,
		Message:   reply,
	}

	_, response := provider.client.CreatePost(post)
	if response.Error != nil {
		log.Printf("unable to post message %+v: %+v", post, response)
		return fmt.Errorf("unable to post message %+v: %+v", post, response)
	}

	return nil
}

func (provider *MessageProvider) BroadcastMessage(message string) error {
	post := &mattermost.Post{
		ChannelId: provider.channel.Id,
		Message:   message,
	}

	_, response := provider.client.CreatePost(post)
	if response.Error != nil {
		log.Printf("unable to post message %+v: %+v", post, response)
		return fmt.Errorf("unable to post message %+v: %+v", post, response)
	}

	return nil
}

func (provider *MessageProvider) GetMessageChannel() chan bot.Message {
	return provider.MessageChannel
}

func (provider *MessageProvider) connect() error {

	protocol := "ws://"
	if provider.Config.Mattermost.Ssl {
		protocol = "wss://"
	}

	connection, appErr := mattermost.NewWebSocketClient(protocol+provider.Config.Mattermost.Server, provider.Config.Mattermost.PrivateAccessToken)
	if appErr != nil {
		return fmt.Errorf("unable to connect to mattermost websocker: %+v", appErr)
	}

	provider.websocketClient = connection

	connection.Listen()

	return nil
}

func (provider *MessageProvider) startReadLoop() {
	log.Println("starting mattermost read loop")
	timeout := provider.Config.Mattermost.ConnectionTimeout

	for {
		provider.websocketClient.Conn.SetWriteDeadline(time.Now().Add(timeout * time.Second))
		provider.websocketClient.Conn.SetReadDeadline(time.Now().Add(timeout * time.Second))

		for event := range provider.websocketClient.EventChannel {
			if event.Event == mattermost.WEBSOCKET_EVENT_POSTED {

				postData, exists := event.Data["post"]
				if !exists {
					log.Printf("invalid postdata from event %+v", event)
					continue
				}

				var jsonString string

				switch t := postData.(type) {
				case string:
					jsonString = postData.(string)
				default:
					log.Printf("invalid data type %s for event %+v", t, event)
					continue
				}

				post := mattermost.PostFromJson(strings.NewReader(jsonString))
				provider.handleMessage(post)
			}
		}
		log.Println("mattermost eventchannel closed. Probably disconnected D:")

		log.Println("Trying to reconnect")

		err := provider.connect()
		if err != nil {
			log.Println("Error trying to connect. Trying again in 10s")
			time.Sleep(10 * time.Second)
			continue
		}

		log.Println("Connected")
	}

}

func (provider *MessageProvider) handleMessage(post *mattermost.Post) {
	channel, response := provider.client.GetChannel(post.ChannelId, "")
	if response.Error != nil {
		log.Printf("unable to get channel by id %s: %+v", post.ChannelId, response)
		return
	}

	// ignore all messaged not from the channel or direct
	if channel.Name != provider.Config.Mattermost.Channel && channel.Type != mattermost.CHANNEL_DIRECT {
		log.Printf("ignoring message from channel %s", channel.Name)
		return
	}

	author, response := provider.client.GetUser(post.UserId, "")
	if response.Error != nil {
		log.Printf("unable to get user by id %s: %+v", post.ChannelId, response)
		return
	}

	msg := bot.Message{
		Target:    post.ChannelId,
		Message:   post.Message,
		IsPrivate: channel.Type == mattermost.CHANNEL_DIRECT,
		Sender: bot.Sender{
			Name:     author.Username,
			NickName: author.Nickname,
		},
	}

	provider.MessageChannel <- msg
}

func (provider *MessageProvider) pingLoop() {
	ticker := time.NewTicker(10 * time.Second)

	timeout := provider.Config.Mattermost.ConnectionTimeout
	log.Printf("Starting ping loop with timeout of %d seconds", timeout)
	provider.websocketClient.Conn.SetWriteDeadline(time.Now().Add(timeout * time.Second))
	provider.websocketClient.Conn.SetReadDeadline(time.Now().Add(timeout * time.Second))

	// push back the timeout by 30 seconds every time we get a pong
	provider.websocketClient.Conn.SetPongHandler(func(appData string) error {
		provider.websocketClient.Conn.SetWriteDeadline(time.Now().Add(timeout * time.Second))
		provider.websocketClient.Conn.SetReadDeadline(time.Now().Add(timeout * time.Second))
		return nil
	})

	for range ticker.C {
		err := provider.websocketClient.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second))
		if err != nil {
			log.Printf("unable to ping: %+v", err)
		}
	}
}

func New(config *bot.Config) *MessageProvider {
	return &MessageProvider{
		MessageChannel: make(chan bot.Message),
		Config:         config,
	}
}
