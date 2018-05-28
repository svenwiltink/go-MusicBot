package bot

import (
	"fmt"
	"strings"

	"github.com/svenwiltink/go-musicbot/music"
	"strconv"
)

type Command struct {
	Name       string
	MasterOnly bool
	Function   func(bot *MusicBot, message Message)
}

var helpCommand = &Command{
	Name: "help",
	Function: func(bot *MusicBot, message Message) {
		helpString := "Available commands: "
		for _, command := range bot.commands {
			helpString += command.Name + " "
		}

		bot.ReplyToMessage(message, helpString)
	},
}

var addCommand = &Command{
	Name: "add",
	Function: func(bot *MusicBot, message Message) {
		words := strings.SplitN(message.Message, " ", 3)
		if len(words) <= 2 {
			bot.ReplyToMessage(message, "No song provided")
			return
		}

		song := &music.Song{
			Path: strings.TrimSpace(words[2]),
		}

		err := bot.musicPlayer.AddSong(song)
		if err != nil {
			bot.ReplyToMessage(message, err.Error())
		} else {
			if message.IsPrivate {
				bot.BroadcastMessage(fmt.Sprintf("%s added by %s", song.Name, message.Sender.NickName))
			}
			bot.ReplyToMessage(message, fmt.Sprintf("%s added", song.Name))
		}
	},
}

var searchAddCommand = &Command{
	Name: "search-add",
	Function: func(bot *MusicBot, message Message) {
		words := strings.SplitN(message.Message, " ", 3)
		if len(words) <= 2 {
			bot.ReplyToMessage(message, "No song provided")
			return
		}

		songs, _ := bot.musicPlayer.Search(words[2])

		if len(songs) == 0 {
			bot.ReplyToMessage(message, "No song found")
			return
		}

		song := songs[0]
		err := bot.musicPlayer.AddSong(song)

		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Error: %v", err))
			return
		}

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("%s added", song.Name))
		}

		bot.ReplyToMessage(message, fmt.Sprintf("%s added", song.Name))
	},
}

var nextCommand = &Command{
	Name: "next",
	Function: func(bot *MusicBot, message Message) {
		err := bot.musicPlayer.Next()
		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Could not skip song: %v", err))
		} else {
			if message.IsPrivate {
				bot.BroadcastMessage("Skipping song")
			}
			bot.ReplyToMessage(message, "Skipping song")
		}
	},
}

var pausedCommand = &Command{
	Name: "pause",
	Function: func(bot *MusicBot, message Message) {
		err := bot.musicPlayer.Pause()
		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Error: %v", err))
			return
		}

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("%s stopped the music", message.Sender.NickName))
		}

		bot.ReplyToMessage(message, "Music paused")

	},
}

var playCommand = &Command{
	Name: "play",
	Function: func(bot *MusicBot, message Message) {
		err := bot.musicPlayer.Play()
		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Error: %v", err))
			return
		}

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("%s resumed the music", message.Sender.NickName))
		}

		bot.ReplyToMessage(message, "Music resumed")

	},
}

var currentCommand = &Command{
	Name: "current",
	Function: func(bot *MusicBot, message Message) {
		song := bot.musicPlayer.GetCurrentSong()
		if song == nil {
			bot.ReplyToMessage(message, "Nothing currently playing")
			return
		}
		bot.ReplyToMessage(message, fmt.Sprintf("Current song: %s %s", song.Artist, song.Name))
	},
}

var whiteListCommand = &Command{
	Name:       "whitelist",
	MasterOnly: true,
	Function: func(bot *MusicBot, message Message) {
		words := strings.SplitN(message.Message, " ", 4)
		if len(words) <= 3 {
			bot.ReplyToMessage(message, "whitelist <add|remove> <name>")
			return
		}

		name := strings.TrimSpace(words[3])
		if len(name) == 0 {
			bot.ReplyToMessage(message, "whitelist <add|remove> <name>")
			return
		}

		if words[2] == "add" {
			err := bot.whitelist.Add(name)
			if err == nil {
				bot.ReplyToMessage(message, fmt.Sprintf("added %s to the whitelist", name))
			} else {
				bot.ReplyToMessage(message, fmt.Sprintf("error: %v", err))
			}
		} else if words[2] == "remove" {
			err := bot.whitelist.Remove(name)
			if err == nil {
				bot.ReplyToMessage(message, fmt.Sprintf("added %s to the whitelist", name))
			} else {
				bot.ReplyToMessage(message, fmt.Sprintf("error: %v", err))
			}
		} else {
			bot.ReplyToMessage(message, "whitelist <add|remove> <name>")
			return
		}
	},
}

var volCommand = &Command{
	Name: "vol",
	Function: func(bot *MusicBot, message Message) {
		words := strings.SplitN(message.Message, " ", 3)

		if len(words) == 2 {
			volume, err := bot.musicPlayer.GetVolume()

			if err != nil {
				bot.ReplyToMessage(message, fmt.Sprintf("unable to get volume: %v", err))
				return
			}

			bot.ReplyToMessage(message, fmt.Sprintf("Current volume: %d", volume))
			return
		}

		// init vars here so we can use them after the switch statement
		volumeString := strings.TrimSpace(words[2])
		var volume int
		var err error

		switch volumeString {
		case "++":
			{
				volume, err = bot.musicPlayer.IncreaseVolume(10)
				if err != nil {
					bot.ReplyToMessage(message, fmt.Sprintf("unable to increase volume: %s", err))
					return
				}
			}
		case "--":
			{
				volume, err = bot.musicPlayer.DecreaseVolume(10)
				if err != nil {
					bot.ReplyToMessage(message, fmt.Sprintf("unable to decrease volume: %s", err))
					return
				}
			}
		default:
			{
				volume, err := strconv.Atoi(strings.TrimSpace(volumeString))

				if err != nil {
					bot.ReplyToMessage(message, fmt.Sprintf("%s is not a valid number", volumeString))
					return
				}

				if volume >= 0 && volume <= 100 {
					bot.musicPlayer.SetVolume(volume)
				} else {
					bot.ReplyToMessage(message, fmt.Sprintf("%s is not a valid volume", volumeString))
					return
				}
			}
		}


		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("Volume set to %d by %s", volume, message.Sender.NickName))
		}

		bot.ReplyToMessage(message, fmt.Sprintf("Volume set to %d", volume))
	},
}

var aboutCommand = &Command{
	Name: "about",
	Function: func(bot *MusicBot, message Message) {
		bot.ReplyToMessage(message, "go-MusicBot by Sven Wiltink: https://github.com/svenwiltink/go-MusicBot")
	},
}