package bot

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/util"
	"os/exec"
	"sort"
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
	Name: "help",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		var names []string
		for commandName := range bot.commands {
			names = append(names, boldText(commandName))
		}
		sort.Strings(names)
		event.Connection.Privmsgf(channel, "Available commands: %s", strings.Join(names, ", "))
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
					message += " " + underlineText(name)
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
	Name: "next",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		_, err := bot.playlist.Next()
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
		}
	},
}

var PlayCommand = Command{
	Name: "play",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		_, err := bot.playlist.Play()
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
		}
	},
}

var PauseCommand = Command{
	Name: "pause",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		err := bot.playlist.Pause()
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
		}
	},
}

var StopCommand = Command{
	Name: "stop",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		err := bot.playlist.Stop()
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
		}
	},
}

var CurrentCommand = Command{
	Name: "current",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		song, remaining := bot.playlist.GetCurrentItem()
		if song != nil {
			event.Connection.Privmsgf(channel, "Current song: %s%s%s "+italicText("(%s remaining)"), BOLD_CHARACTER, formatSong(song), BOLD_CHARACTER, util.FormatSongLength(remaining))
		} else {
			event.Connection.Privmsg(channel, italicText("Nothing currently playing"))
		}
	},
}

var AddCommand = Command{
	Name: "add",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		if len(parameters) < 1 {
			channel := event.Arguments[0]
			event.Connection.Privmsg(channel, boldText("!music add <music link>"))
			return
		}
		channel := event.Arguments[0]
		url := parameters[0]

		items, err := bot.playlist.AddItems(url)
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
		} else {
			var songs []string
			i := 8
			for _, item := range items {
				songs = append(songs, formatSong(item))
				i--
				if i < 0 {
					songs = append(songs, italicText(fmt.Sprintf("and %d more..", len(items)-8)))
					break
				}
			}
			event.Connection.Privmsgf(channel, "%s added song(s): %s", event.Nick, strings.Join(songs, " | "))
		}
		bot.playlist.Play()
	},
}

var OpenCommand = Command{
	Name: "open",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		if len(parameters) < 1 {
			channel := event.Arguments[0]
			event.Connection.Privmsg(channel, boldText("!music open <music link>"))
			return
		}
		channel := event.Arguments[0]
		url := parameters[0]

		items, err := bot.playlist.InsertItems(url, 0)
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
		} else {
			var songs []string
			i := 8
			for _, item := range items {
				songs = append(songs, formatSong(item))
				i--
				if i < 0 {
					songs = append(songs, italicText(fmt.Sprintf("and %d more..", len(items)-8)))
					break
				}
			}
			event.Connection.Privmsgf(channel, "%s added song(s): %s", event.Nick, strings.Join(songs, " | "))
		}
		bot.playlist.Next()
	},
}

var ShuffleCommand = Command{
	Name: "shuffle",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		bot.playlist.ShuffleList()
		event.Connection.Privmsg(channel, italicText("The playlist has been shuffled"))
	},
}

var ListCommand = Command{
	Name: "list",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		items := bot.playlist.GetItems()
		if len(items) == 0 {
			event.Connection.Privmsg(channel, italicText("The playlist is empty"))
		}

		for i, item := range items {
			event.Connection.Privmsgf(channel, "%d. %s", i+1, formatSong(item))

			if i >= 9 && len(items) > 10 {
				event.Connection.Privmsgf(channel, italicText("And %d more.."), len(items)-10)
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
		event.Connection.Privmsg(channel, italicText("The playlist is now empty"))
	},
}

var SearchCommand = Command{
	Name: "search",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		if len(parameters) < 1 {
			event.Connection.Privmsg(channel, "!music search [<playerName>] <search term>")
			return
		}

		results, err := searchSongs(bot.playlist, parameters)
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
			return
		}
		if len(results) == 0 {
			event.Connection.Privmsg(channel, italicText("Nothing found!"))
			return
		}
		for plyr, res := range results {
			for i, item := range res {
				event.Connection.Privmsgf(channel, "[%s #%d] %s - %s", plyr, i+1, formatSong(item), item.GetURL())
			}
		}
	},
}

var SearchAddCommand = Command{
	Name: "search-add",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		if len(parameters) < 1 {
			event.Connection.Privmsg(channel, "!music search-add [<playerName>] <search term>")
			return
		}

		results, err := searchSongs(bot.playlist, parameters)
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
			return
		}
		if len(results) == 0 {
			event.Connection.Privmsg(channel, italicText("Nothing found!"))
			return
		}
		for plyr, res := range results {
			for _, item := range res {
				event.Connection.Privmsgf(channel, "%s added song: %s (%s)", event.Nick, formatSong(item), italicText(plyr))
				_, err := bot.playlist.AddItems(item.GetURL())
				if err != nil {
					event.Connection.Privmsg(channel, inverseText(err.Error()))
					return
				}
				bot.playlist.Play()
				return
			}
		}
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
