package bot

import (
	"fmt"
	irc "github.com/thoj/go-ircevent"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/playlist"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type MusicBot struct {
	Commands      map[string]Command
	Whitelist     []string
	playlist      playlist.ListInterface
	Configuration *config.IRC
}

func NewMusicBot(conf *config.IRC, playlst playlist.ListInterface) (mb *MusicBot, err error) {
	whitelist, err := config.ReadWhitelist(conf.WhiteListPath)
	if err != nil {
		return
	}

	mb = &MusicBot{
		Commands:      make(map[string]Command),
		Whitelist:     whitelist,
		Configuration: conf,
		playlist:      playlst,
	}
	return
}

func (m *MusicBot) getCommand(name string) (command Command, exists bool) {
	command, exists = m.Commands[name]
	return
}

func (m *MusicBot) registerCommand(command Command) bool {
	if _, exists := m.getCommand(command.Name); !exists {
		m.Commands[strings.ToLower(command.Name)] = command
		fmt.Println("registered the " + command.Name + " command")
		return true
	}
	return false
}

func (m *MusicBot) isUserWhitelisted(realname string) (iswhitelisted bool, index int) {
	for index, name := range m.Whitelist {
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
	m.registerCommand(PauseCommand)
	m.registerCommand(CurrentCommand)
	m.registerCommand(ShuffleCommand)
	m.registerCommand(ListCommand)
	m.registerCommand(FlushCommand)
	m.registerCommand(OpenCommand)
	m.registerCommand(SearchCommand)

	m.registerCommand(VolUpCommand)
	m.registerCommand(VolDownCommand)
	m.registerCommand(VolCommand)

	irccon := irc.IRC(m.Configuration.Nick, m.Configuration.Realname)
	irccon.Password = m.Configuration.Password
	irccon.UseTLS = m.Configuration.Ssl

	err = irccon.Connect(m.Configuration.Server)
	if err != nil {
		return
	}

	irccon.AddCallback("001", func(event *irc.Event) {
		event.Connection.Join(m.Configuration.Channel)
	})
	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		channel := event.Arguments[0]
		message := event.Arguments[len(event.Arguments)-1]
		realname := event.User

		if strings.HasPrefix(message, "!music") {
			if isWhiteListed, _ := m.isUserWhitelisted(realname); m.Configuration.Master == realname || isWhiteListed {
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
			}
		}
	})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	m.playlist.Stop()
}
