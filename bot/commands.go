package bot

import (
	"fmt"
	"github.com/thoj/go-ircevent"
	"os/exec"
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
		for commandName := range bot.Commands {
			message += " " + commandName
		}

		event.Connection.Privmsg(channel, message)
	},
}

var WhitelistCommand = Command{
	Name: "Whitelist",
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
				for _, name := range bot.Whitelist {
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
				if realname == bot.Master {
					if isWhitelisted, _ := bot.isUserWhitelisted(user); !isWhitelisted {
						bot.Whitelist = append(bot.Whitelist, user)
						event.Connection.Privmsg(channel, "User added to whitelist")
						writeWhitelist(bot.Whitelist)
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
		bot.MusicPlayer.Next()
	},
}

var PlayCommand = Command{
	Name: "Play",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		bot.MusicPlayer.Pause()
	},
}

var PauseCommand = Command{
	Name: "Pause",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		bot.MusicPlayer.Pause()
	},
}

var CurrentCommand = Command{
	Name: "Current",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		url := bot.MusicPlayer.CurrentSong.GetURL()
		message := fmt.Sprintf("Current song: %s", url)
		event.Connection.Privmsg(channel, message)
	},
}

var OpenCommand = Command{
	Name: "Open",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		if len(parameters) < 1 {
			channel := event.Arguments[0]
			event.Connection.Privmsg(channel, "!music open <youtube link>")
			return
		}

		url := parameters[0]
		fmt.Println(url)
		bot.MusicPlayer.AddSong(url)
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
		message := fmt.Sprint("Shuffeling queue")
		bot.MusicPlayer.Queue.Shuffle()
		event.Connection.Privmsg(channel, message)
	},
}

var ListCommand = Command{
	Name: "list",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]
		for i, item := range bot.MusicPlayer.Queue.Items {
			message := fmt.Sprintf("#%d %s", i, item.GetURL())
			event.Connection.Privmsg(channel, message)
		}
	},
}

var FlushCommand = Command{
	Name: "flush",
	Function: func(bot *MusicBot, event *irc.Event, parameters []string) {
		channel := event.Arguments[0]

		message := fmt.Sprint("Flushing queue")
		bot.MusicPlayer.Queue.Flush()
		event.Connection.Privmsg(channel, message)
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
