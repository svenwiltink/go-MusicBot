package bot

import (
	"fmt"
	"log"
	"strings"

	"github.com/svenwiltink/go-musicbot/pkg/music"
	"github.com/svenwiltink/go-musicbot/pkg/music/dataprovider/m3u"
	"github.com/svenwiltink/go-musicbot/pkg/music/dataprovider/nts"
	"github.com/svenwiltink/go-musicbot/pkg/music/dataprovider/soundcloud"
	"github.com/svenwiltink/go-musicbot/pkg/music/dataprovider/youtube"
	"github.com/svenwiltink/go-musicbot/pkg/music/player"
	"github.com/svenwiltink/go-musicbot/pkg/music/provider/mpv"
)

type MusicBot struct {
	messageProvider MessageProvider
	musicPlayer     music.Player
	config          *Config
	commands        map[string]*Command
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
				m3u.DataProvider{},
				youtubeProvider,
			},
		),
		commands: make(map[string]*Command),
	}

	return instance
}

func (bot *MusicBot) Start() {

	bot.loadWhitelist()
	bot.musicPlayer.Start()
	bot.registerCommands()

	bot.musicPlayer.AddListener(music.EventSongStarted, func(arguments ...interface{}) {
		song := arguments[0].(*music.Song)
		bot.BroadcastMessage(fmt.Sprintf("Started playing %s: %s", song.Artist, song.Name))
	})

	bot.musicPlayer.AddListener(music.EventSongStartError, func(arguments ...interface{}) {
		song := arguments[0].(*music.Song)
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
	bot.registerCommand(searchAddCommand)
	bot.registerCommand(nextCommand)
	bot.registerCommand(pausedCommand)
	bot.registerCommand(playCommand)
	bot.registerCommand(currentCommand)
	bot.registerCommand(queueCommand)
	bot.registerCommand(flushCommand)
	bot.registerCommand(whiteListCommand)
	bot.registerCommand(volCommand)
	bot.registerCommand(aboutCommand)
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

func (bot *MusicBot) BroadcastMessage(message string) {
	bot.messageProvider.BroadcastMessage(message)
}

func (bot *MusicBot) Stop() {
	bot.musicPlayer.Stop()
}

func (bot *MusicBot) handleMessage(message Message) {
	if strings.HasPrefix(message.Message, bot.config.CommandPrefix) {
		if !(bot.config.Master == message.Sender.Name || bot.whitelist.Contains(message.Sender.Name)) {
			bot.ReplyToMessage(message, fmt.Sprintf("You're not on the whitelist %s", message.Sender.Name))
			return
		}

		words := strings.SplitN(message.Message, " ", 3)
		if len(words) >= 2 {
			word := strings.TrimSpace(words[1])
			command := bot.getCommand(word)
			if command == nil {
				bot.ReplyToMessage(message, fmt.Sprintf("Unknown command. Use %s help for help", bot.config.CommandPrefix))
				return
			}

			if command.MasterOnly && bot.config.Master != message.Sender.Name {
				bot.ReplyToMessage(message, "this command is for masters only")
				return
			}

			command.Function(bot, message)
		} else {
			bot.ReplyToMessage(message, fmt.Sprintf("Use %s help to list all the commands", bot.config.CommandPrefix))
		}
	}
}

func (bot *MusicBot) GetMusicPlayer() music.Player {
	return bot.musicPlayer
}
