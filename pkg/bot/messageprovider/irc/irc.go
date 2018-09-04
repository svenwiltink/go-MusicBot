package irc

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"

	ircclient "github.com/fluffle/goirc/client"
	"github.com/svenwiltink/go-musicbot/pkg/bot"
)

type MessageProvider struct {
	Config         *bot.Config
	MessageChannel chan bot.Message
	IrcConnection  *ircclient.Conn
}

func (irc *MessageProvider) Start() error {
	ircConfig := ircclient.NewConfig(irc.Config.Irc.Nick, irc.Config.Irc.RealName)
	ircConfig.Server = irc.Config.Irc.Server
	ircConfig.Pass = irc.Config.Irc.Pass
	ircConfig.Timeout = time.Second * 5

	log.Printf("%+v", irc.Config.Irc)
	if irc.Config.Irc.Ssl {
		ircConfig.SSL = true
		ircConfig.SSLConfig = &tls.Config{ServerName: strings.Split(irc.Config.Irc.Server, ":")[0]}
	}

	irc.IrcConnection = ircclient.Client(ircConfig)
	irc.IrcConnection.HandleFunc(ircclient.CONNECTED, func(conn *ircclient.Conn, line *ircclient.Line) {
		log.Println("joining channel")
		conn.Join(irc.Config.Irc.Channel)
	})

	irc.IrcConnection.HandleFunc(ircclient.PRIVMSG, func(conn *ircclient.Conn, line *ircclient.Line) {
		log.Printf("ident: %v", line.Ident)
		log.Printf("message: %s", line.Text())

		irc.MessageChannel <- bot.Message{
			Message:   line.Text(),
			Target:    line.Target(),
			IsPrivate: line.Target() != irc.Config.Irc.Channel,
			Sender: bot.Sender{
				Name:     line.Ident,
				NickName: line.Nick,
			},
		}
	})

	irc.IrcConnection.HandleFunc(ircclient.ERROR, func(conn *ircclient.Conn, line *ircclient.Line) {
		log.Printf("IRC error %v", line)
	})

	log.Printf("Trying to connect to server %s", irc.Config.Irc.Server)
	err := irc.IrcConnection.Connect()

	if err != nil {
		return fmt.Errorf("could not start client: %v", err)

	}

	log.Printf("connected")

	return nil
}

func (irc *MessageProvider) SendReplyToMessage(message bot.Message, reply string) error {
	irc.IrcConnection.Privmsg(message.Target, reply)
	return nil
}

func (irc *MessageProvider) BroadcastMessage(message string) error {
	irc.IrcConnection.Privmsg(irc.Config.Irc.Channel, message)
	return nil
}

func (irc *MessageProvider) GetMessageChannel() chan bot.Message {
	return irc.MessageChannel
}

func New(config *bot.Config) *MessageProvider {
	return &MessageProvider{
		MessageChannel: make(chan bot.Message),
		Config:         config,
	}
}
