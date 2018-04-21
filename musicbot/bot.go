package musicbot

type MusicBot struct {
	messageProvider MessageProvider
	config          *Config
}

func NewMusicBot(config *Config, messageProvider MessageProvider) *MusicBot {
	instance := &MusicBot{
		config:          config,
		messageProvider: messageProvider,
	}

	return instance
}

func (bot *MusicBot) Start() {
	for message := range bot.messageProvider.GetMessageChannel() {
		bot.messageProvider.SendReplyToMessage(message, message.Message)
	}
}
