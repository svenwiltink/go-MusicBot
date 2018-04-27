package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/svenwiltink/go-musicbot/music"
	"github.com/svenwiltink/go-musicbot/music/player"
	"github.com/svenwiltink/go-musicbot/music/provider/mpv"
)

type MusicBot struct {
	messageProvider MessageProvider
	musicPlayer     music.Player
	config          *Config
	commands        map[string]*Command
}

func NewMusicBot(config *Config, messageProvider MessageProvider) *MusicBot {

	mpvPlayer := mpv.NewPlayer(config.MpvPath, config.MpvSocket)
	err := mpvPlayer.Start()

	if err != nil {
		log.Printf("unable to start music player: %v", err)
		return nil
	}

	instance := &MusicBot{
		config:          config,
		messageProvider: messageProvider,
		musicPlayer:     player.NewMusicPlayer(mpvPlayer),
		commands:        make(map[string]*Command),
	}

	return instance
}

func (bot *MusicBot) Start() {
	bot.musicPlayer.Start()
	bot.registerCommands()

	go bot.messageLoop()
}

func (bot *MusicBot) messageLoop() {
	for message := range bot.messageProvider.GetMessageChannel() {
		if strings.HasPrefix(message.Message, bot.config.CommandPrefix) {
			words := strings.SplitN(message.Message, " ", 3)
			if len(words) >= 2 {
				word := strings.TrimSpace(words[1])
				command := bot.getCommand(word)
				if command != nil {
					command.Function(bot, message)
				}
			} else {
				bot.ReplyToMessage(message, fmt.Sprintf("Use %s help to list all the commands", bot.config.CommandPrefix))
			}
		}
	}
}

func (bot *MusicBot) registerCommands() {
	bot.registerCommand(HelpCommand)
	bot.registerCommand(AddCommand)
	bot.registerCommand(NextCommand)
}

func (bot *MusicBot) registerCommand(command *Command) {
	bot.commands[command.Name] = command
}

func (bot *MusicBot) getCommand(name string) *Command {
	command, _ := bot.commands[name]
	return command
}

func (bot *MusicBot) ReplyToMessage(message Message, reply string) {
	bot.messageProvider.SendReplyToMessage(message, reply)
}

func (bot *MusicBot) Stop() {
	bot.musicPlayer.Stop()
}
