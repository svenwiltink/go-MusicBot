package bot

type MessageProvider interface {
	GetMessageChannel() chan Message
	SendReplyToMessage(message Message, reply string) error
	Start() error
}
