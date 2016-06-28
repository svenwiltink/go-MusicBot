package main

import (
	"github.com/thoj/go-ircevent"
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
					event.Connection.Privmsg(channel, "User removed from whitelist")
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