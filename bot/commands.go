package bot

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/util"
	"os/exec"
	"strings"
)

type Command struct {
	Name     string
	Function func(bot *MusicBot, event *irc.Event, parameters []string)
}

func (c *Command) execute(bot *MusicBot, event *irc.Event, parameters []string) {
	c.Function(bot, event, parameters)
}

var HelpCommand = Command{
	Name: "Help",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		message := "Available commands: "
		for commandName := range bot.commands {
			message += " " + commandName
		}

		event.Connection.Privmsg(channel, message)
	},
}

var WhitelistCommand = Command{
	Name: "whitelist",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		realname := event.User
		if len(parameters) < 1 {
			event.Connection.Privmsg(channel, "!music whitelist <show|add|remove> [user]")
			return
		}

		subcommand := parameters[0]
		switch subcommand {
		case "show":
			{
				message := "Current whitelist: "
				for _, name := range bot.whitelist {
					message += " " + name
				}

				event.Connection.Privmsg(channel, message)
			}
		case "add":
			{
				if len(parameters) < 2 {
					event.Connection.Privmsg(channel, "!music whitelist add [user]")
					return
				}
				user := parameters[1]
				if realname == bot.conf.Master {
					if isWhitelisted, _ := bot.isUserWhitelisted(user); !isWhitelisted {
						bot.whitelist = append(bot.whitelist, user)

						err := config.WriteWhitelist(bot.conf.WhiteListPath, bot.whitelist)
						if err != nil {
							event.Connection.Privmsg(channel, err.Error())
							return
						}
						event.Connection.Privmsgf(channel, "User %s added to whitelist", user)
					}
				}
			}
		case "remove":
			{
				if len(parameters) < 2 {
					event.Connection.Privmsg(channel, "!music whitelist remove [user]")
					return
				}
				user := parameters[1]
				if realname == bot.conf.Master {
					if isWhitelisted, index := bot.isUserWhitelisted(user); isWhitelisted {
						bot.whitelist = append(bot.whitelist[:index], bot.whitelist[index+1:]...)

						err := config.WriteWhitelist(bot.conf.WhiteListPath, bot.whitelist)
						if err != nil {
							event.Connection.Privmsg(channel, err.Error())
							return
						}
						event.Connection.Privmsgf(channel, "User %s removed from whitelist", user)
					}
				}
			}
		}
	},
}

var NextCommand = Command{
	Name: "Next",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		_, err := bot.playlist.Next()
		if err != nil {
			event.Connection.Privmsg(channel, err.Error())
		}
	},
}

var PlayCommand = Command{
	Name: "Play",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		_, err := bot.playlist.Play()
		if err != nil {
			event.Connection.Privmsg(channel, err.Error())
		}
	},
}

var PauseCommand = Command{
	Name: "Pause",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		err := bot.playlist.Pause()
		if err != nil {
			event.Connection.Privmsg(channel, err.Error())
		}
	},
}

var CurrentCommand = Command{
	Name: "Current",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		song, remaining := bot.playlist.GetCurrentItem()
		title := "Not playing"
		if song != nil {
			title = song.GetTitle()
		}

		event.Connection.Privmsg(channel, fmt.Sprintf("Current song: %s [%s] (%s remaining)", title, util.FormatSongLength(song.GetDuration()), util.FormatSongLength(remaining)))
	},
}

var OpenCommand = Command{
	Name: "add",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		if len(parameters) < 1 {
			channel := event.Arguments[0]
			event.Connection.Privmsg(channel, "!music open <music link>")
			return
		}
		channel := event.Arguments[0]
		url := parameters[0]

		items, err := bot.playlist.AddItems(url)
		if err != nil {
			event.Connection.Privmsg(channel, err.Error())
		} else {
			var songs []string
			i := 8
			for _, item := range items {
				songs = append(songs, item.GetTitle())
				i--
				if i < 0 {
					songs = append(songs, fmt.Sprintf("(and %d more)", len(items)-8))
					break
				}
			}
			event.Connection.Privmsgf(channel, "%s added song(s): %s", event.Nick, strings.Join(songs, ", "))
		}
		bot.playlist.Play()
	},
}

var SearchCommand = Command{
	Name: "search",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		if len(parameters) < 1 {
			event.Connection.Privmsg(channel, "!music search <search term>")
			return
		}
		event.Connection.Privmsg(channel, "Not implemented")
	},
}

var ShuffleCommand = Command{
	Name: "shuffle",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		bot.playlist.ShuffleList()
		event.Connection.Privmsg(channel, "Shuffling queue")
	},
}

var ListCommand = Command{
	Name: "list",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		items := bot.playlist.GetItems()
		for i, item := range items {
			event.Connection.Privmsgf(channel, "%d. %s [%s]", i+1, item.GetTitle(), util.FormatSongLength(item.GetDuration()))

			if i > 10 && len(items) > 12 {
				event.Connection.Privmsgf(channel, "And %d more..", len(items)-10)
				return
			}
		}
	},
}

var FlushCommand = Command{
	Name: "flush",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]

		bot.playlist.EmptyList()
		event.Connection.Privmsg(channel, fmt.Sprint("Flushing queue"))
	},
}

var VolUpCommand = Command{
	Name: "vol++",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		cmd := exec.Command("amixer", "-D", "pulse", "sset", "Master", "10%+")
		cmd.Run()
	},
}

var VolDownCommand = Command{
	Name: "vol--",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		cmd := exec.Command("amixer", "-D", "pulse", "sset", "Master", "10%-")
		cmd.Run()
	},
}

var VolCommand = Command{
	Name: "vol",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		if len(parameters) < 1 {
			channel := event.Arguments[0]
			event.Connection.Privmsg(channel, "!music vol <volume>")
			return
		}
		cmd := exec.Command("amixer", "-D", "pulse", "sset", "Master", parameters[0]+"%")
		cmd.Run()
	},
}
