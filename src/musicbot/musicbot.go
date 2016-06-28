package main

import (
	irc "github.com/thoj/go-ircevent"
	"os"
	"strings"
	"fmt"
	"encoding/json"
)

type MusicBot struct {
	Commands map[string]Command
}

func NewMusicBot() *MusicBot {
	return &MusicBot{Commands:make(map[string]Command)}
}

func(m *MusicBot) getCommand(name string) (command Command, exists bool) {
	command, exists = m.Commands[name]
	return command, exists
}

func(m *MusicBot) registerCommand(command Command) bool {
	if _, exists := m.getCommand(command.Name); !exists {
		m.Commands[command.Name] = command
		fmt.Println("registered the " + command.Name + " command")
		return true
	}
	return false
}

type Configuration struct {
	Server 		string
	Ssl 		bool
	Channel		string
	Realname 	string
	Nick 		string
	Password 	string
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

	bot := NewMusicBot()
	bot.registerCommand(HelpCommand)

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

		fmt.Println(channel)
		fmt.Println(realname)
		fmt.Println(message)

		if strings.HasPrefix(message, "!music") {
			fmt.Println("music prefix found")
			command, _ := bot.getCommand("Help")
			command.execute([]string{"what", "up"})
		}
	})

	irccon.Wait()
}
