package terminal

import (
	"bufio"
	"github.com/svenwiltink/go-musicbot/musicbot"
	"log"
	"os"
)

type MessageProvider struct {
	channel chan musicbot.Message
}

func (messageProvider *MessageProvider) GetMessageChannel() chan musicbot.Message {
	return messageProvider.channel
}

func (messageProvider *MessageProvider) SendReplyToMessage(message musicbot.Message, reply string) error {
	log.Printf(reply)
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
		messageProvider.channel <- musicbot.Message{
			Message: text,
			Target:  "#test",
			Sender: musicbot.Sender{
				Name:     "terminal",
				NickName: "terminal",
			},
		}
	}
}

func New() *MessageProvider {
	return &MessageProvider{
		channel: make(chan musicbot.Message),
	}
}
