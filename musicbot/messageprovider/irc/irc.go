package irc

import (
	"fmt"
	ircclient "github.com/fluffle/goirc/client"
	"github.com/svenwiltink/go-musicbot/musicbot"
	"log"
	"crypto/tls"
	"strings"
)

type MessageProvider struct {
	Config         *musicbot.Config
	MessageChannel chan musicbot.Message
	IrcConnection  *ircclient.Conn
}

func (irc *MessageProvider) Start() error {
	ircConfig := ircclient.NewConfig(irc.Config.Irc.Nick, irc.Config.Irc.RealName)
	ircConfig.Server = irc.Config.Irc.Server
	ircConfig.Pass = irc.Config.Irc.Pass

	log.Printf("%+v", irc.Config.Irc)
	if irc.Config.Irc.Ssl {
		log.Println("enabling ssl")
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

		irc.MessageChannel <- musicbot.Message{
			Message: line.Text(),
			Target:  line.Target(),
			Sender: musicbot.Sender{
				Name:     line.Ident,
				NickName: line.Nick,
			},
		}
	})

	log.Printf("Trying to connect to server %s", irc.Config.Irc.Server)
	err := irc.IrcConnection.Connect()

	if err != nil {
		return fmt.Errorf("could not start client: %v", err)

	}

	log.Printf("connected")

	return nil
}

func (irc *MessageProvider) SendReplyToMessage(message musicbot.Message, reply string) error {
	irc.IrcConnection.Privmsg(message.Target, reply)
	return nil
}

func (irc *MessageProvider) GetMessageChannel() chan musicbot.Message {
	return irc.MessageChannel
}

func New(config *musicbot.Config) *MessageProvider {
	return &MessageProvider{
		MessageChannel: make(chan musicbot.Message),
		Config: config,
	}
}
