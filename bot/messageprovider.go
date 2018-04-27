package bot

type MessageProvider interface {
	GetMessageChannel() chan Message
	SendReplyToMessage(message Message, reply string) error
	BroadcastMessage(message string) error
	Start() error
}
