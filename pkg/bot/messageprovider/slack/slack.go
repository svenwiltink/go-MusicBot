package slack

import (
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"github.com/svenwiltink/go-musicbot/pkg/bot"
	"log"
	"os"
)

type MessageProvider struct {
	Config         *bot.Config
	MessageChannel chan bot.Message

	rtm *socketmode.Client
	api *slack.Client
}

func (provider *MessageProvider) Start() error {
	provider.api = slack.New(provider.Config.Slack.Token, slack.OptionAppLevelToken(provider.Config.Slack.ApplicationToken), slack.OptionDebug(true), slack.OptionLog(log.New(os.Stderr, "slack-bot", log.Lshortfile|log.LstdFlags)))
	log.Println(provider.Config.Slack.Token)

	provider.rtm = socketmode.New(provider.api)
	go func() {
		err := provider.rtm.Run()
		provider.rtm.Run()
		if err != nil {
			panic(err)
		}
	}()

	go provider.handleMessages()

	return nil
}

func (provider *MessageProvider) SendReplyToMessage(message bot.Message, reply string) error {
	parameters := slack.NewPostMessageParameters()
	parameters.LinkNames = 1
	_, _, _, err := provider.api.SendMessage(message.Target, slack.MsgOptionPostMessageParameters(parameters), slack.MsgOptionText(reply, false))
	return err
}

func (provider *MessageProvider) BroadcastMessage(message string) error {
	parameters := slack.NewPostMessageParameters()
	parameters.LinkNames = 1
	_, _, _, err := provider.api.SendMessage(provider.Config.Slack.Channel, slack.MsgOptionPostMessageParameters(parameters), slack.MsgOptionText(message, false))
	return err
}

func (provider *MessageProvider) GetMessageChannel() chan bot.Message {
	return provider.MessageChannel
}

func (provider *MessageProvider) handleMessages() {
	for msg := range provider.rtm.Events {
		log.Printf("Event Received: %#v", msg)

		switch msg.Type {
		case socketmode.EventTypeEventsAPI:

			provider.rtm.Ack(*msg.Request)

			event := msg.Data.(slackevents.EventsAPIEvent)

			switch event.Type {
			case slackevents.CallbackEvent:
				innerEvent := event.InnerEvent
				switch ev := innerEvent.Data.(type) {
				case *slackevents.MessageEvent:
					log.Println("user: ", ev.User)
					log.Println("username: ", ev.Username)
					log.Println("message: ", ev.Text)
					log.Println("channel: ", ev.Channel)

					usermention := fmt.Sprintf("<@%s>", ev.User)

					message := bot.Message{
						Message: ev.Text,
						Sender: bot.Sender{
							Name:     usermention,
							NickName: usermention,
						},
						Target:    ev.Channel,
						IsPrivate: ev.Channel != provider.Config.Slack.Channel,
					}

					go func() {
						provider.MessageChannel <- message
					}()
				}

			default:
				log.Printf("unsupported Events API event received %T", event.InnerEvent.Data)
			}
		default:
			log.Printf("ignoreing event type %s", msg.Type)
		}
	}
}

func New(config *bot.Config) *MessageProvider {
	return &MessageProvider{
		MessageChannel: make(chan bot.Message),
		Config:         config,
	}
}
