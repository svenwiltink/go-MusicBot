package dummy

import (
	"log"
	"time"

	"github.com/svenwiltink/go-musicbot/musicbot"
)

type MessageProvider struct {
	channel chan musicbot.Message
}

func (messageProvider *MessageProvider) GetMessageChannel() chan musicbot.Message {
	return messageProvider.channel
}

func (messageProvider *MessageProvider) SendReplyToMessage(message musicbot.Message, reply string) error {
	log.Println("sending reply")
	return nil
}

func (messageProvider *MessageProvider) Start() error {
	go func() {
		for true {
			messageProvider.channel <- musicbot.Message{
				Message: "Dummy message",
				Target:  "#test",
				Sender: musicbot.Sender{
					Name:     "fakeUser",
					NickName: "Actually fake",
				},
			}

			time.Sleep(time.Second * 1)
		}
	}()

	return nil
}

func New() *MessageProvider {
	return &MessageProvider{
		channel: make(chan musicbot.Message),
	}
}
