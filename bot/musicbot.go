package bot

import (
	"bufio"
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

func NewMusicBot(conf *config.IRC, playlst playlist.ListInterface) *MusicBot {
	whitelist, err := readWhitelist()
	if err != nil {
		println("Error: " + err.Error())
	}

	return &MusicBot{
		Commands:      make(map[string]Command),
		Whitelist:     whitelist,
		Configuration: conf,
		playlist:      playlst,
	}
}

func (m *MusicBot) getCommand(name string) (command Command, exists bool) {
	command, exists = m.Commands[name]
	return command, exists
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

func (m *MusicBot) Start() {
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

	err := irccon.Connect(m.Configuration.Server)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
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

func readWhitelist() ([]string, error) {
	file, err := os.Open("whitelist.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func writeWhitelist(lines []string) error {
	file, err := os.Create("whitelist.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
