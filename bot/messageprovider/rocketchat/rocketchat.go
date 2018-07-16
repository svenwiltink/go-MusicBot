package rocketchat

import (
	"fmt"
	"github.com/svenwiltink/go-musicbot/bot"
	"github.com/svenwiltink/rocketchatgo"
	"log"
)

type MessageProvider struct {
	session    *rocketchatgo.Session
	msgChannel chan bot.Message
	Config     *bot.Config
}

func (p *MessageProvider) GetMessageChannel() chan bot.Message {
	return p.msgChannel
}

func (p *MessageProvider) SendReplyToMessage(message bot.Message, reply string) error {
	return p.session.SendMessage(message.Target, reply)
}

func (p *MessageProvider) BroadcastMessage(message string) error {
	channel := p.session.GetChannelByName(p.Config.Rocketchat.Channel)

	if channel == nil {
		return fmt.Errorf("unable to find channel %s", p.Config.Rocketchat.Channel)
	}

	return p.session.SendMessage(channel.ID, message)
}

func (p *MessageProvider) Start() error {
	session, err := rocketchatgo.NewClient(p.Config.Rocketchat.Server, p.Config.Rocketchat.Ssl)

	if err != nil {
		return err
	}

	p.session = session
	err = p.session.Login(p.Config.Rocketchat.Username, "", p.Config.Rocketchat.Pass)

	if err != nil {
		return err
	}

	p.session.AddHandler(p.onMessageCreate)

	return err
}

func (p *MessageProvider) onMessageCreate(session *rocketchatgo.Session, event *rocketchatgo.MessageCreateEvent) {
	if session.GetUserID() == event.Message.Sender.ID {
		return
	}

	channel := p.session.GetChannelById(event.Message.ChannelID)

	if channel == nil {
		log.Printf("Unable to find channel with id %s, for message %+v", event.Message.ChannelID, event.Message)
		return
	}

	msg := bot.Message{
		Target:    event.Message.ChannelID,
		Message:   event.Message.Message,
		IsPrivate: channel.Type == rocketchatgo.RoomTypeDirect,
		Sender: bot.Sender{
			Name:     event.Message.Sender.Username,
			NickName: event.Message.Sender.Name,
		},
	}

	p.msgChannel <- msg

}

func New(config *bot.Config) *MessageProvider {
	return &MessageProvider{
		msgChannel: make(chan bot.Message),
		Config:     config,
	}
}
