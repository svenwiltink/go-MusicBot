package musicbot

import (
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
			Name: words[2],
			Path: words[2],
		}

		err := bot.musicPlayer.AddSong(song)
		if err != nil {
			bot.ReplyToMessage(message, err.Error())
		} else {
			bot.ReplyToMessage(message, "song added")
		}
	},
}
