package bot

import (
	"github.com/thoj/go-ircevent"
	"gitlab.transip.us/swiltink/go-MusicBot/config"
	"gitlab.transip.us/swiltink/go-MusicBot/player"
	"gitlab.transip.us/swiltink/go-MusicBot/util"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
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
		target, _, _ := bot.getTarget(event)
		var names []string
		for commandName := range bot.commands {
			names = append(names, boldText(commandName))
		}
		sort.Strings(names)
		event.Connection.Privmsgf(target, "Available commands: %s", strings.Join(names, ", "))
	},
}

var WhitelistCommand = Command{
	Name: "whitelist",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		realname := event.User
		if len(parameters) < 1 {
			event.Connection.Privmsg(target, "Usage: !music whitelist <show|add|remove> [user]")
			return
		}

		subcommand := parameters[0]
		switch subcommand {
		case "show":
			message := "Current whitelist: "
			for _, name := range bot.whitelist {
				message += " " + underlineText(name)
			}
			event.Connection.Privmsg(target, message)
		case "add":
			if len(parameters) < 2 {
				event.Connection.Privmsg(target, boldText("Usage: !music whitelist add [user]"))
				return
			}
			user := parameters[1]
			if realname == bot.conf.Master {
				if isWhitelisted, _ := bot.isUserWhitelisted(user); !isWhitelisted {
					bot.whitelist = append(bot.whitelist, user)

					err := config.WriteWhitelist(bot.conf.WhiteListPath, bot.whitelist)
					if err != nil {
						event.Connection.Privmsg(target, err.Error())
						return
					}
					event.Connection.Privmsgf(target, boldText("User %s added to whitelist by %s"), user, event.Nick)
				}
			}
		case "remove":
			if len(parameters) < 2 {
				event.Connection.Privmsg(target, boldText("Usage: !music whitelist remove [user]"))
				return
			}
			user := parameters[1]
			if realname == bot.conf.Master {
				if isWhitelisted, index := bot.isUserWhitelisted(user); isWhitelisted {
					bot.whitelist = append(bot.whitelist[:index], bot.whitelist[index+1:]...)

					err := config.WriteWhitelist(bot.conf.WhiteListPath, bot.whitelist)
					if err != nil {
						event.Connection.Privmsg(target, err.Error())
						return
					}
					event.Connection.Privmsgf(target, boldText("User %s removed from whitelist by %s"), user, event.Nick)
				}
			}
		}
	},
}

var NextCommand = Command{
	Name: "next",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		_, err := bot.player.Next()
		if err != nil {
			event.Connection.Privmsg(target, inverseText(err.Error()))
			return
		}
		bot.announceMessage(true, event, boldText(event.Nick)+" skipped the song")
	},
}

var PlayCommand = Command{
	Name: "play",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		_, err := bot.player.Play()
		if err != nil {
			event.Connection.Privmsg(target, inverseText(err.Error()))
			return
		}
		bot.announceMessage(true, event, boldText(event.Nick)+" started the player")
	},
}

var SeekCommand = Command{
	Name: "seek",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		if len(parameters) < 1 {
			event.Connection.Privmsg(target, boldText("Usage: !music seek <secondsInSong> Or: !music seek <percentage>%"))
			return
		}
		seekStr := parameters[0]
		var seekSeconds int64
		if strings.HasSuffix(seekStr, "%") {
			percentage, err := strconv.ParseInt(seekStr[:len(seekStr)-1], 10, 32)
			if err != nil {
				event.Connection.Privmsg(target, boldText("Error parsing seek percentage"))
				return
			}
			song, _ := bot.player.GetCurrentSong()
			duration := song.GetDuration().Nanoseconds() / 100 * percentage
			seekSeconds = duration / int64(time.Second)
		} else {
			var err error
			seekSeconds, err = strconv.ParseInt(seekStr, 10, 32)
			if err != nil {
				event.Connection.Privmsg(target, boldText("Error parsing seek seconds"))
				return
			}
		}
		err := bot.player.Seek(int(seekSeconds))
		if err != nil {
			event.Connection.Privmsg(target, inverseText(err.Error()))
			return
		}
	},
}

var PauseCommand = Command{
	Name: "pause",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		err := bot.player.Pause()
		if err != nil {
			event.Connection.Privmsg(target, inverseText(err.Error()))
			return
		}
		state := bot.player.GetStatus()
		switch state {
		case player.PAUSED:
			bot.announceMessage(false, event, boldText(event.Nick)+" paused the player")
		case player.PLAYING:
			bot.announceMessage(false, event, boldText(event.Nick)+" unpaused the player")
		}
	},
}

var StopCommand = Command{
	Name: "stop",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		err := bot.player.Stop()
		if err != nil {
			event.Connection.Privmsg(target, inverseText(err.Error()))
		}
		bot.announceMessage(true, event, boldText(event.Nick)+" stopped the player")
	},
}

var CurrentCommand = Command{
	Name: "current",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		song, remaining := bot.player.GetCurrentSong()
		if song != nil {
			event.Connection.Privmsgf(target, "Current song: %s%s%s "+italicText("(%s remaining)"), BOLD_CHARACTER, formatSong(song), BOLD_CHARACTER, util.FormatSongLength(remaining))
		} else {
			event.Connection.Privmsg(target, italicText("Nothing currently playing"))
		}
	},
}

var AddCommand = Command{
	Name: "add",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		if len(parameters) < 1 {
			event.Connection.Privmsg(target, boldText("!music add <music link>"))
			return
		}
		url := parameters[0]

		songs, err := bot.player.AddSongs(url)
		if err != nil {
			event.Connection.Privmsg(target, inverseText(err.Error()))
			return
		}
		bot.announceAddedSongs(event, songs)
		bot.player.Play()
	},
}

var OpenCommand = Command{
	Name: "open",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		if len(parameters) < 1 {
			event.Connection.Privmsg(target, boldText("Usage: !music open <music link>"))
			return
		}
		url := parameters[0]

		songs, err := bot.player.InsertSongs(url, 0)
		if err != nil {
			event.Connection.Privmsg(target, inverseText(err.Error()))
			return
		}
		bot.announceAddedSongs(event, songs)
		bot.player.Next()
	},
}

var ShuffleCommand = Command{
	Name: "shuffle",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		bot.player.ShuffleQueue()
		bot.announceMessage(true, event, boldText(event.Nick)+" shuffled the playlist")
	},
}

var ListCommand = Command{
	Name: "list",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		items := bot.player.GetQueuedSongs()
		if len(items) == 0 {
			event.Connection.Privmsg(target, italicText("The playlist is empty"))
		}

		for i, item := range items {
			event.Connection.Privmsgf(target, "%d. %s", i+1, formatSong(item))

			if i >= 9 && len(items) > 10 {
				event.Connection.Privmsgf(target, italicText("And %d more.."), len(items)-10)
				return
			}
		}
	},
}

var FlushCommand = Command{
	Name: "flush",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		bot.player.EmptyQueue()
		bot.announceMessage(true, event, boldText(event.Nick)+" emptied the playlist")
	},
}

var SearchCommand = Command{
	Name: "search",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		if len(parameters) < 1 {
			event.Connection.Privmsg(target, boldText("Usage: !music search [<playerName>] <search term>"))
			return
		}

		results, err := searchSongs(bot.player, parameters)
		if err != nil {
			event.Connection.Privmsg(target, inverseText(err.Error()))
			return
		}
		if len(results) == 0 {
			event.Connection.Privmsg(target, italicText("Nothing found!"))
			return
		}
		for plyr, res := range results {
			for i, item := range res {
				event.Connection.Privmsgf(target, "[%s #%d] %s - %s", plyr, i+1, formatSong(item), item.GetURL())
			}
		}
	},
}

var SearchAddCommand = Command{
	Name: "search-add",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		target, _, _ := bot.getTarget(event)
		if len(parameters) < 1 {
			event.Connection.Privmsg(target, boldText("Usage: !music search-add [<playerName>] <search term>"))
			return
		}

		results, err := searchSongs(bot.player, parameters)
		if err != nil {
			event.Connection.Privmsg(target, inverseText(err.Error()))
			return
		}
		if len(results) == 0 {
			event.Connection.Privmsg(target, italicText("Nothing found!"))
			return
		}
		for plyr, res := range results {
			for _, item := range res {
				bot.announceMessagef(false, event, "%s added song: %s (%s)", boldText(event.Nick), formatSong(item), italicText(plyr))
				_, err := bot.player.AddSongs(item.GetURL())
				if err != nil {
					event.Connection.Privmsg(target, inverseText(err.Error()))
					return
				}
				bot.player.Play()
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
		target, _, _ := bot.getTarget(event)
		if len(parameters) < 1 {
			event.Connection.Privmsg(target, "!music vol <volume>")
			return
		}
		cmd := exec.Command("amixer", "-D", "pulse", "sset", "Master", parameters[0]+"%")
		cmd.Run()
	},
}
