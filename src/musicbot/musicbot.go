package main

import (
	irc "github.com/thoj/go-ircevent"
	"os"
	"strings"
	"fmt"
	"encoding/json"
	"bufio"
)

type MusicBot struct {
	Commands 	map[string]Command
	Whitelist 	[]string
	Master		string
}

func NewMusicBot(c Configuration) *MusicBot {
	whitelist, err := readWhitelist()
	if err != nil {
		println("Error: " + err.Error())
	}
	return &MusicBot{
		Commands: make(map[string]Command),
		Whitelist: whitelist,
		Master: c.Master,
	}
}

func(m *MusicBot) getCommand(name string) (command Command, exists bool) {
	command, exists = m.Commands[name]
	return command, exists
}

func(m *MusicBot) registerCommand(command Command) bool {
	if _, exists := m.getCommand(command.Name); !exists {
		m.Commands[strings.ToLower(command.Name)] = command
		fmt.Println("registered the " + command.Name + " command")
		return true
	}
	return false
}

func(m *MusicBot) isUserWhitelisted(realname string) (iswhitelisted bool, index int) {
	for index, name := range m.Whitelist {
		if name == realname {
			return true, index
		}
	}
	return false, -1
}

type Configuration struct {
	Server 		string
	Ssl 		bool
	Channel		string
	Realname 	string
	Nick 		string
	Password 	string
	Master		string
}

func main() {

	file, err := os.Open("conf.json")
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(2)
	}

	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(2)
	}

	bot := NewMusicBot(configuration)
	bot.registerCommand(HelpCommand)
	bot.registerCommand(WhitelistCommand)
	bot.registerCommand(NextCommand)
	bot.registerCommand(PlayCommand)
	bot.registerCommand(PauseCommand)
	bot.registerCommand(CurrentCommand)
	bot.registerCommand(OpenCommand)
	bot.registerCommand(SearchCommand)

	bot.registerCommand(VolUpCommand)
	bot.registerCommand(VolDownCommand)
	bot.registerCommand(VolCommand)

	irccon := irc.IRC(configuration.Nick, configuration.Realname)
	irccon.Password = configuration.Password
	irccon.UseTLS = configuration.Ssl
	irccon.Debug = true

	err = irccon.Connect(configuration.Server)

	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	irccon.AddCallback("001", func(event *irc.Event){
		event.Connection.Join(configuration.Channel)
	})

	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		channel := event.Arguments[0]
		message := event.Arguments[len(event.Arguments)-1]
		realname := event.User

		if strings.HasPrefix(message, "!music") {
			if isWhiteListed, _ := bot.isUserWhitelisted(realname); bot.Master == realname || isWhiteListed {
				arguments := strings.Split(message, " ")[1:]
				if len(arguments) > 0 {
					commandName := strings.ToLower(arguments[0])
					arguments = arguments[1:]
					if command, exist := bot.getCommand(commandName); exist {
						command.execute(bot, event, arguments)
						return
					}
				}
				event.Connection.Privmsg(channel, "Unknown command. Use !music help to list all the commands available")
			}
		}
	})

	irccon.Wait()
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
