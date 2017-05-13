package bot

import (
	"fmt"
	irc "github.com/thoj/go-ircevent"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"gitlab.transip.us/swiltink/go-MusicBot/songplayer"
	"strings"
)

type MusicBot struct {
	ircConn   *irc.Connection
	commands  map[string]Command
	whitelist []string
	player    player.MusicPlayer
	conf      *config.IRC
}

func NewMusicBot(conf *config.IRC, player player.MusicPlayer) (mb *MusicBot, err error) {
	whitelist, err := config.ReadWhitelist(conf.WhiteListPath)
	if err != nil {
		return
	}

	mb = &MusicBot{
		commands:  make(map[string]Command),
		whitelist: whitelist,
		conf:      conf,
		player:    player,
	}
	return
}

func (m *MusicBot) getCommand(name string) (command Command, exists bool) {
	command, exists = m.commands[name]
	return
}

func (m *MusicBot) registerCommand(command Command) bool {
	if _, exists := m.getCommand(command.Name); !exists {
		m.commands[strings.ToLower(command.Name)] = command
		fmt.Println("registered the " + command.Name + " command")
		return true
	}
	return false
}

func (m *MusicBot) isUserWhitelisted(realname string) (iswhitelisted bool, index int) {
	for index, name := range m.whitelist {
		if name == realname {
			return true, index
		}
	}
	return false, -1
}

func (m *MusicBot) Start() (err error) {
	m.registerCommand(HelpCommand)
	m.registerCommand(WhitelistCommand)

	m.registerCommand(NextCommand)
	m.registerCommand(PlayCommand)
	m.registerCommand(SeekCommand)
	m.registerCommand(PauseCommand)
	m.registerCommand(StopCommand)

	m.registerCommand(CurrentCommand)
	m.registerCommand(ShuffleCommand)
	m.registerCommand(ListCommand)
	m.registerCommand(FlushCommand)

	m.registerCommand(AddCommand)
	m.registerCommand(OpenCommand)

	m.registerCommand(SearchCommand)
	m.registerCommand(SearchAddCommand)

	m.registerCommand(VolUpCommand)
	m.registerCommand(VolDownCommand)
	m.registerCommand(VolCommand)

	m.ircConn = irc.IRC(m.conf.Nick, m.conf.Realname)
	m.ircConn.Password = m.conf.Password
	m.ircConn.UseTLS = m.conf.Ssl

	err = m.ircConn.Connect(m.conf.Server)
	if err != nil {
		return
	}

	m.ircConn.AddCallback("001", func(event *irc.Event) {
		event.Connection.Join(m.conf.Channel)
	})
	m.ircConn.AddCallback("PRIVMSG", func(event *irc.Event) {
		channel := event.Arguments[0]
		message := event.Arguments[len(event.Arguments)-1]
		realname := event.User

		if strings.HasPrefix(message, "!music") {
			if isWhiteListed, _ := m.isUserWhitelisted(realname); m.conf.Master == realname || isWhiteListed {
				arguments := strings.Split(message, " ")[1:]
				if len(arguments) > 0 {
					commandName := strings.ToLower(arguments[0])
					arguments = arguments[1:]
					if command, exist := m.getCommand(commandName); exist {
						command.execute(m, event, arguments)
						return
					}
				}
				event.Connection.Privmsg(channel, "Unknown command. Use !music help to list all the commands available")
			} else {
				event.Connection.Privmsgf(channel, "I will not obey you, %s", realname)
			}
		}
	})

	m.player.AddListener("play_start", m.onPlay)
	return
}

func (m *MusicBot) Stop() (err error) {
	err = m.player.Stop()
	return
}

func (m *MusicBot) getTarget(event *irc.Event) (target string, isPrivate, isMain bool) {
	if len(event.Arguments) == 0 {
		return
	}
	target = event.Arguments[0]
	if !strings.HasPrefix(target, "#") {
		target = event.Nick
		isPrivate = true
	}
	isMain = target == m.conf.Channel
	return
}

func (m *MusicBot) announceAddedSongs(event *irc.Event, songs []songplayer.Playable) {
	var songTitles []string
	i := 6
	for _, song := range songs {
		songTitles = append(songTitles, formatSong(song))
		i--
		if i < 0 {
			songTitles = append(songTitles, italicText(fmt.Sprintf("and %d more..", len(songs)-6)))
			break
		}
	}
	m.announceMessagef(false, event, "%s added song(s): %s", boldText(event.Nick), strings.Join(songTitles, " | "))
}

func (m *MusicBot) Announce(message string) {
	m.ircConn.Privmsg(m.conf.Channel, message)
}

func (m *MusicBot) announceMessage(nonMainOnly bool, event *irc.Event, message string) {
	target, isPrivate, isMain := m.getTarget(event)
	if isPrivate || !nonMainOnly {
		event.Connection.Privmsg(target, message)
	}
	if isPrivate || (!isMain && !nonMainOnly) {
		// Announce it to the main channel as well
		event.Connection.Privmsg(m.conf.Channel, message)
	}
}

func (m *MusicBot) announceMessagef(nonMainOnly bool, event *irc.Event, format string, a ...interface{}) {
	target, isPrivate, isMain := m.getTarget(event)
	if isPrivate || !nonMainOnly {
		event.Connection.Privmsgf(target, format, a...)
	}
	if isPrivate || (!isMain && !nonMainOnly) {
		// Announce it to the main channel as well
		event.Connection.Privmsgf(m.conf.Channel, format, a...)
	}
}

func (m *MusicBot) onPlay(args ...interface{}) {
	if len(args) < 1 {
		return
	}

	itm, ok := args[0].(songplayer.Playable)
	if !ok {
		return
	}

	m.ircConn.Actionf(m.conf.Channel, "starts playing: %s", boldText(formatSong(itm)))
}
