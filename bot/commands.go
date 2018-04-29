package bot

import (
	"fmt"
	"strings"

	"github.com/svenwiltink/go-musicbot/music"
)

type Command struct {
	Name     string
	Function func(bot *MusicBot, message Message)
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
				bot.BroadcastMessage(fmt.Sprintf("%s added", song.Name))
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
