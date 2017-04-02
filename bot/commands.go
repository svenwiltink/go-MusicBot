package bot

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/util"
	"os/exec"
	"strings"
	"sort"
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
	Name: "Next",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		_, err := bot.playlist.Next()
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
		}
	},
}

var PlayCommand = Command{
	Name: "Play",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		_, err := bot.playlist.Play()
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
		}
	},
}

var PauseCommand = Command{
	Name: "Pause",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		err := bot.playlist.Pause()
		if err != nil {
			event.Connection.Privmsg(channel, inverseText(err.Error()))
		}
	},
}

var CurrentCommand = Command{
	Name: "Current",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		song, remaining := bot.playlist.GetCurrentItem()
		if song != nil {
			event.Connection.Privmsgf(channel, "Current song: %s "+italicText("(%s remaining)"), formatSong(song), util.FormatSongLength(remaining))
		} else {
			event.Connection.Privmsg(channel, italicText("Nothing currently playing"))
		}
	},
}

var OpenCommand = Command{
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
					songs = append(songs, italicText(fmt.Sprintf("(and %d more)", len(items)-8)))
					break
				}
			}
			event.Connection.Privmsgf(channel, "%s added song(s): %s", event.Nick, strings.Join(songs, ", "))
		}
		bot.playlist.Play()
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

			if i > 10 {
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
			event.Connection.Privmsg(channel, "!music search <search term>")
			return
		}
		searchStr := strings.Join(parameters, " ")
		resultsFound := false
		for _, p := range bot.playlist.GetPlayers() {
			items, err := p.SearchItems(searchStr, 3)
			if err != nil {
				event.Connection.Privmsg(channel, inverseText(err.Error()))
				continue
			}
			for i, item := range items {
				resultsFound = true
				event.Connection.Privmsgf(channel, "[%s #%d] %s - %s", p.Name(), i+1, formatSong(&item), item.GetURL())
			}
		}
		if !resultsFound {
			event.Connection.Privmsg(channel, inverseText("Nothing found!"))
		}
	},
}

var SearchAddCommand = Command{
	Name: "search-add",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		if len(parameters) < 1 {
			event.Connection.Privmsg(channel, "!music search-add <search term>")
			return
		}
		searchStr := strings.Join(parameters, " ")
		resultsFound := false
		for _, p := range bot.playlist.GetPlayers() {
			items, err := p.SearchItems(searchStr, 1)
			if err != nil {
				event.Connection.Privmsg(channel, inverseText(err.Error()))
				continue
			}
			for _, item := range items {
				resultsFound = true
				_, err := bot.playlist.AddItems(item.GetURL())
				if err != nil {
					event.Connection.Privmsg(channel, inverseText(err.Error()))
					continue
				}
				event.Connection.Privmsgf(channel, "%s added song: %s", event.Nick, formatSong(&item))
			}
		}
		if !resultsFound {
			event.Connection.Privmsg(channel, inverseText("Nothing found!"))
		}
		bot.playlist.Play()
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
