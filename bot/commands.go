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

var HelpCommand = &Command{
	Name: "help",
	Function: func(bot *MusicBot, message Message) {
		bot.ReplyToMessage(message, "this is a help command")
	},
}

var AddCommand = &Command{
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

var SearchAddCommand = &Command{
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
		} else {
			if message.IsPrivate {
				bot.BroadcastMessage(fmt.Sprintf("%s added", song.Name))
			}
			bot.ReplyToMessage(message, fmt.Sprintf("%s added", song.Name))
		}
	},
}

var NextCommand = &Command{
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

var CurrentCommand = &Command{
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

var WhiteListCommand = &Command{
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

var VolCommand = &Command{
	Name: "vol",
	Function: func(bot *MusicBot, message Message) {
		words := strings.SplitN(message.Message, " ", 3)
		if len(words) <= 2 {
			bot.ReplyToMessage(message, "vol <volume>")
			return
		}

		volume, err := strconv.Atoi(strings.TrimSpace(words[2]))

		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("%s is not a valid number", words[2]))
			return
		}

		if volume >= 0 && volume <= 100 {
			bot.musicPlayer.SetVolume(volume)
		} else {
			bot.ReplyToMessage(message, fmt.Sprintf("%s is not a valid volume", words[2]))
			return
		}

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("Volume set to %d by %s", volume, message.Sender.NickName))
		}

		bot.ReplyToMessage(message, fmt.Sprintf("Volume set to %d", volume))
	},
}
