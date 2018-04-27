package terminal

import (
	"bufio"
	"log"
	"os"

	"github.com/svenwiltink/go-musicbot/bot"
)

type MessageProvider struct {
	channel chan bot.Message
}

func (messageProvider *MessageProvider) GetMessageChannel() chan bot.Message {
	return messageProvider.channel
}

func (messageProvider *MessageProvider) SendReplyToMessage(message bot.Message, reply string) error {
	log.Printf(reply)
	return nil
}

func (messageProvider *MessageProvider) BroadcastMessage(message string) error {
	log.Printf(message)
	return nil
}

func (messageProvider *MessageProvider) Start() error {
	go messageProvider.start()

	return nil
}

func (messageProvider *MessageProvider) start() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		messageProvider.channel <- bot.Message{
			Message:   text,
			IsPrivate: false,
			Target:    "#test",
			Sender: bot.Sender{
				Name:     "terminal",
				NickName: "terminal",
			},
		}
	}
}

func New() *MessageProvider {
	return &MessageProvider{
		channel: make(chan bot.Message),
	}
}
