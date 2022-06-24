package bot

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/svenwiltink/go-musicbot/pkg/music"
	"github.com/svenwiltink/go-musicbot/pkg/music/dataprovider/nts"
	"github.com/svenwiltink/go-musicbot/pkg/music/dataprovider/soundcloud"
	"github.com/svenwiltink/go-musicbot/pkg/music/dataprovider/youtube"
	"github.com/svenwiltink/go-musicbot/pkg/music/player"
	"github.com/svenwiltink/go-musicbot/pkg/music/provider/mpv"
)

var (
	errCommandNotFound  = errors.New("command not found")
	errVariableNotFound = errors.New("command variable not found")

	Version   = "placeholder string"
	GoVersion = "placeholder string"
	BuildDate = "placeholder string"
)

type MusicBot struct {
	messageProvider MessageProvider
	musicPlayer     music.Player
	config          *Config
	commands        map[string]Command
	commandAliases  map[string]Command
	whitelist       *WhiteList
}

func NewMusicBot(config *Config, messageProvider MessageProvider) *MusicBot {

	mpvPlayer := mpv.NewPlayer(config.MpvPath, config.MpvSocket)
	err := mpvPlayer.Start()

	if err != nil {
		log.Printf("unable to start music player: %v", err)
		return nil
	}

	youtubeProvider, err := youtube.NewDataProvider(config.Youtube.APIKey)

	if err != nil {
		log.Printf("unable to start youtube provider: %v", err)
		return nil
	}

	instance := &MusicBot{
		config:          config,
		messageProvider: messageProvider,
		musicPlayer: player.NewMusicPlayer(
			[]music.Provider{mpvPlayer},
			[]music.DataProvider{
				nts.DataProvider{},
				soundcloud.DataProvider{},
				youtubeProvider,
			},
		),
		commands:       make(map[string]Command),
		commandAliases: make(map[string]Command),
	}

	return instance
}

func (bot *MusicBot) Start() {

	bot.loadWhitelist()
	bot.musicPlayer.Start()
	bot.registerCommands()

	bot.musicPlayer.AddListener(music.EventSongStarted, func(arguments ...interface{}) {
		song := arguments[0].(music.Song)
		bot.BroadcastMessage(fmt.Sprintf("Started playing %s: %s", song.Artist, song.Name))
	})

	bot.musicPlayer.AddListener(music.EventSongStartError, func(arguments ...interface{}) {
		song := arguments[0].(music.Song)
		bot.BroadcastMessage(fmt.Sprintf("Error starting %v %v, skipping", song.Artist, song.Name))
	})

	go bot.messageLoop()
}

func (bot *MusicBot) loadWhitelist() {
	whitelist, err := LoadWhiteList(bot.config.WhiteListFile)

	if err != nil {
		log.Println(err)
		bot.whitelist = &WhiteList{
			names: make(map[string]struct{}, 0),
		}

		return
	}

	bot.whitelist = whitelist
}

func (bot *MusicBot) messageLoop() {
	for message := range bot.messageProvider.GetMessageChannel() {
		bot.handleMessage(message)
	}
}

func (bot *MusicBot) registerCommands() {
	bot.registerCommand(helpCommand)
	bot.registerCommand(addCommand)
	bot.registerCommand(searchCommand)
	bot.registerCommand(searchAddCommand)
	bot.registerCommand(nextCommand)
	bot.registerCommand(pausedCommand)
	bot.registerCommand(playCommand)
	bot.registerCommand(currentCommand)
	bot.registerCommand(queueCommand)
	bot.registerCommand(flushCommand)
	bot.registerCommand(shuffleCommand)
	bot.registerCommand(whiteListCommand)
	bot.registerCommand(volCommand)
	bot.registerCommand(aboutCommand)
}

func (bot *MusicBot) registerCommand(command Command) {
	bot.commands[command.Name] = command
	for _, alias := range command.Aliases {
		bot.commandAliases[alias] = command
	}
}

// getCommand returns the command by name or an error if it could not be found
func (bot *MusicBot) getCommand(name string) (Command, error) {
	command, exists := bot.commands[name]
	if exists {
		return command, nil
	}
	command, exists = bot.commandAliases[name]
	if exists {
		return command, nil
	}
	return Command{}, errCommandNotFound
}

func (bot *MusicBot) ReplyToMessage(message Message, reply string) {
	bot.messageProvider.SendReplyToMessage(message, reply)
}

func (bot *MusicBot) BroadcastMessage(message string) {
	bot.messageProvider.BroadcastMessage(message)
}

func (bot *MusicBot) Stop() {
	bot.musicPlayer.Stop()
}

func (bot *MusicBot) handleMessage(message Message) {
	if strings.HasPrefix(message.Message, bot.config.CommandPrefix+" ") {
		message.Message = strings.TrimPrefix(message.Message, bot.config.CommandPrefix+" ")
		bot.handleCommand(message)
		return
	}
	if strings.HasPrefix(message.Message, bot.config.ShortCommandPrefix) {
		// either with or without space after the ShortPrefix is fine
		message.Message = strings.TrimSpace(strings.TrimPrefix(message.Message, bot.config.ShortCommandPrefix))
		bot.handleCommand(message)
	}
}

func (bot *MusicBot) handleCommand(message Message) {
	if !(bot.config.Master == message.Sender.Name || bot.whitelist.Contains(message.Sender.Name)) {
		bot.ReplyToMessage(message, fmt.Sprintf("You're not on the whitelist %s", message.Sender.Name))
		return
	}

	words := strings.SplitN(message.Message, " ", 2)
	if len(words) >= 1 {
		commandWord := Message.getCommandWord(message)
		command, err := bot.getCommand(commandWord)
		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Unknown command. Use %s help for help", bot.config.CommandPrefix))
			return
		}

		if command.MasterOnly && bot.config.Master != message.Sender.Name {
			bot.ReplyToMessage(message, "This command is for masters only")
			return
		}

		command.Function(bot, message)
	} else {
		bot.ReplyToMessage(message, fmt.Sprintf("Use %s help to list all the commands", bot.config.CommandPrefix))
	}
}

func (bot *MusicBot) GetMusicPlayer() music.Player {
	return bot.musicPlayer
}
