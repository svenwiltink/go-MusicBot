package main

import (
	"github.com/thoj/go-ircevent"
	"os/exec"
	"strings"
	"fmt"
)

type Command struct {
	Name string
	Function func(bot *MusicBot, event *irc.Event, parameters []string)
}

func(c *Command) execute(bot *MusicBot, event *irc.Event, parameters []string) {
	c.Function(bot, event, parameters)
}

var HelpCommand = Command {
	Name:"Help",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		message := "Available commands: "
		for commandName,_ := range bot.Commands {
			message += " " + commandName
		}

		event.Connection.Privmsg(channel, message)
	},
}

var WhitelistCommand = Command {
	Name: "Whitelist",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		realname := event.User
		if len(parameters) < 1 {
			event.Connection.Privmsg(channel, "!music whitelist <show|add|remove> [user]")
			return;
		}

		subcommand := parameters[0]
		switch subcommand{
		case "show": {
			message := "Current whitelist: "
			for _, name := range bot.Whitelist {
				message += " " + name
			}

			event.Connection.Privmsg(channel, message)
		}
		case "add": {
			if len(parameters) < 2 {
				event.Connection.Privmsg(channel, "!music whitelist add [user]")
				return
			}
			user := parameters[1]
			if realname == bot.Master {
				if isWhitelisted, _ := bot.isUserWhitelisted(user); !isWhitelisted {
					bot.Whitelist = append(bot.Whitelist, user)
					event.Connection.Privmsg(channel, "User added to whitelist")
					writeWhitelist(bot.Whitelist)
				}
			}
		}
		case "remove": {
			if len(parameters) < 2 {
				event.Connection.Privmsg(channel, "!music whitelist remove [user]")
				return
			}
			user := parameters[1]
			if realname == bot.Master {
				if isWhitelisted, index := bot.isUserWhitelisted(user); isWhitelisted {
					bot.Whitelist = append(bot.Whitelist[:index], bot.Whitelist[index+1:]...)
					event.Connection.Privmsg(channel, "User removed from whitelist")
					writeWhitelist(bot.Whitelist)
				}
			}
		}
		}
	},
}

var NextCommand = Command{
	Name: "Next",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		cmd := exec.Command("./bin/spotify.sh", "next")
		cmd.Run()
	},
}

var PlayCommand = Command {
	Name: "Play",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		cmd := exec.Command("./bin/spotify.sh", "play")
		cmd.Run()
	},
}

var PauseCommand = Command {
	Name: "Pause",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		cmd := exec.Command("./bin/spotify.sh", "pause")
		cmd.Run()
	},
}

var CurrentCommand = Command {
	Name: "Current",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		cmd := exec.Command("./bin/spotify.sh", "current")
		result, _ := cmd.Output()
		resultString := string(result)
		resultArray := strings.Split(resultString, "\n")
		message := strings.Join(append(resultArray[0:1], resultArray[2:len(resultArray) - 1]...), " | ")
		message = strings.Replace(message, "    ",  " ", -1)
		message = strings.Replace(message, "   ",  " ", -1)
		event.Connection.Privmsg(channel, message)
	},
}

var OpenCommand = Command {
	Name: "Open",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		if len(parameters) < 1 {
			channel := event.Arguments[0]
			event.Connection.Privmsg(channel, "!music open <spotify link>")
			return
		}
		spotifyUri := parameters[0]
		fmt.Println(spotifyUri)
		cmd := exec.Command("./bin/spotify.sh", "open",  spotifyUri )
		cmd.Run()
	},
}

var SearchCommand = Command {
	Name: "search",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		if len(parameters) < 1 {
			channel := event.Arguments[0]
			event.Connection.Privmsg(channel, "!music search <search term>")
			return
		}
		cmd := exec.Command("./bin/spotify.sh", "search",  strings.Join(parameters, " "))
		cmd.Run()
	},
}

var VolUpCommand = Command {
	Name: "vol++",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		cmd := exec.Command("amixer", "-D", "pulse", "sset", "Master", "10%+")
		cmd.Run()
	},
}

var VolDownCommand = Command {
	Name: "vol--",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		cmd := exec.Command("amixer", "-D", "pulse", "sset", "Master", "10%-")
		cmd.Run()
	},
}

var VolCommand = Command {
	Name: "vol",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		if len(parameters) < 1 {
			channel := event.Arguments[0]
			event.Connection.Privmsg(channel, "!music vol <volume>")
			return
		}
		cmd := exec.Command("amixer", "-D", "pulse", "sset", "Master", parameters[0] + "%")
		cmd.Run()
	},
}
