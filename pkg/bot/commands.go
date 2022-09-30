package bot

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/svenwiltink/go-musicbot/pkg/music"
)

type Command struct {
	Name      string
	Aliases   []string
	AdminOnly bool
	Function  func(bot *MusicBot, message Message)
}

var helpCommand = Command{
	Name:    "help",
	Aliases: []string{"h"},
	Function: func(bot *MusicBot, message Message) {
		helpString := "Available commands: "
		for _, command := range bot.commands {
			helpString += command.Name
			if len(command.Aliases) > 0 {
				helpString += "[" + strings.Join(command.Aliases, ", ") + "]"
			}
			helpString += " "
		}

		bot.ReplyToMessage(message, helpString)
	},
}

func sanitizeSong(song string) string {
	song = strings.TrimSpace(song)
	song = strings.TrimLeft(song, "<")
	song = strings.TrimRight(song, ">")
	return song
}

var addCommand = Command{
	Name:    "add",
	Aliases: []string{"a"},
	Function: func(bot *MusicBot, message Message) {
		parameter, cmdParamError := message.getCommandParameter()
		if cmdParamError != nil {
			bot.ReplyToMessage(message, "No song provided")
			return
		}

		song := music.Song{
			Path: sanitizeSong(parameter),
		}

		song, err := bot.musicPlayer.AddSong(song)
		if err != nil {
			bot.ReplyToMessage(message, err.Error())
			return
		}

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("%s: %s added by %s", song.Artist, song.Name, message.Sender.Name))
		}
		bot.ReplyToMessage(message, fmt.Sprintf("%s: %s added", song.Artist, song.Name))

	},
}

var searchCommand = Command{
	Name:    "search",
	Aliases: []string{"s"},
	Function: func(bot *MusicBot, message Message) {
		parameter, cmdParamError := message.getCommandParameter()
		if cmdParamError != nil {
			bot.ReplyToMessage(message, "No song provided")
			return
		}

		songs, err := bot.musicPlayer.Search(parameter)

		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("error: %v", err))
		}

		if len(songs) == 0 {
			bot.ReplyToMessage(message, "No song found")
			return
		}

		var builder strings.Builder

		for number, song := range songs {
			builder.WriteString(fmt.Sprintf("%d  %s - %s (%s)\n", number+1, song.Artist, song.Name, song.Duration))
		}

		bot.ReplyToMessage(message, builder.String())

	},
}

var searchAddCommand = Command{
	Name:    "search-add",
	Aliases: []string{"sa"},
	Function: func(bot *MusicBot, message Message) {
		parameter, cmdParamError := message.getCommandParameter()
		if cmdParamError != nil {
			bot.ReplyToMessage(message, "No song provided")
			return
		}

		songs, err := bot.musicPlayer.Search(parameter)

		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("error: %v", err))
		}

		if len(songs) == 0 {
			bot.ReplyToMessage(message, "No song found")
			return
		}

		song := songs[0]
		song, err = bot.musicPlayer.AddSong(song)

		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Error: %v", err))
			return
		}

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("%s: %s added by %s", song.Artist, song.Name, message.Sender.Name))
		}

		bot.ReplyToMessage(message, fmt.Sprintf("%s: %s added", song.Artist, song.Name))
	},
}

var nextCommand = Command{
	Name:    "next",
	Aliases: []string{"n"},
	Function: func(bot *MusicBot, message Message) {
		err := bot.musicPlayer.Next()
		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Could not skip song: %v", err))
		} else {
			if message.IsPrivate {
				bot.BroadcastMessage(fmt.Sprintf("%s skipped the song", message.Sender.Name))
			}
			bot.ReplyToMessage(message, "Skipping song")
		}
	},
}

var pausedCommand = Command{
	Name:    "pause",
	Aliases: []string{},
	Function: func(bot *MusicBot, message Message) {
		err := bot.musicPlayer.Pause()
		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Error: %v", err))
			return
		}

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("%s stopped the music", message.Sender.Name))
		}

		bot.ReplyToMessage(message, "Music paused")

	},
}

var playCommand = Command{
	Name:    "play",
	Aliases: []string{},
	Function: func(bot *MusicBot, message Message) {
		err := bot.musicPlayer.Play()
		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Error: %v", err))
			return
		}

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("%s resumed the music", message.Sender.Name))
		}

		bot.ReplyToMessage(message, "Music resumed")

	},
}

var currentCommand = Command{
	Name:    "current",
	Aliases: []string{"c"},
	Function: func(bot *MusicBot, message Message) {
		song, durationLeft := bot.musicPlayer.GetCurrentSong()
		if song == nil {
			bot.ReplyToMessage(message, "Nothing currently playing")
			return
		}

		if song.SongType == music.SongTypeSong {
			bot.ReplyToMessage(
				message,
				fmt.Sprintf("Current song: %s %s. %s remaining (%s)", song.Artist, song.Name, durationLeft.String(), song.Duration.Round(time.Second).String()))
		} else {
			bot.ReplyToMessage(
				message,
				fmt.Sprintf("Current song: %s %s. This is a livestream, use the next command to skip", song.Artist, song.Name))
		}
	},
}

var queueCommand = Command{
	Name:    "queue",
	Aliases: []string{"q"},
	Function: func(bot *MusicBot, message Message) {
		queue := bot.GetMusicPlayer().GetQueue()

		queueLength := queue.GetLength()
		nextSongs, _ := queue.GetNextN(5)
		duration := queue.GetTotalDuration()

		bot.ReplyToMessage(message, fmt.Sprintf("%d songs in the queue. Total duration %s", queueLength, duration.String()))

		for index, song := range nextSongs {
			bot.ReplyToMessage(message, fmt.Sprintf("#%d, %s: %s (%s)\n", index+1, song.Artist, song.Name, song.Duration.String()))
		}

		if queueLength > 5 {
			bot.ReplyToMessage(message, fmt.Sprintf("and %d more", queueLength-5))
		}

	},
}

var queueDeleteCommand = Command{
	Name: "queue-delete",
	Function: func(bot *MusicBot, message Message) {
		// !music queue-delete #
		parameter, cmdParamError := message.getCommandParameter()
		if cmdParamError != nil {
			bot.ReplyToMessage(message, "No queue item provided")
			return
		}

		queueItem, err := strconv.Atoi(parameter)
		if err != nil {
			bot.ReplyToMessage(message, "Invalid queue-item provided")
			return
		}

		queue := bot.GetMusicPlayer().GetQueue()
		if err = queue.Delete(queueItem - 1); err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("Could not delete queue-item: %s", err))
			return
		}

		bot.ReplyToMessage(message, fmt.Sprintf("queue-item %d deleted", queueItem))
	},
}

var flushCommand = Command{
	Name:    "flush",
	Aliases: []string{"f"},
	Function: func(bot *MusicBot, message Message) {
		bot.musicPlayer.GetQueue().Flush()

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("%s flushed the queue", message.Sender.Name))
		}

		bot.ReplyToMessage(message, "Queue flushed")
	},
}

var shuffleCommand = Command{
	Name:    "shuffle",
	Aliases: []string{},
	Function: func(bot *MusicBot, message Message) {
		bot.musicPlayer.GetQueue().Shuffle()

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("%s shuffled the queue", message.Sender.Name))
		}

		bot.ReplyToMessage(message, "Queue shuffled")
	},
}

var allowListCommand = Command{
	Name:      "allowlist",
	Aliases:   []string{},
	AdminOnly: true,
	Function: func(bot *MusicBot, message Message) {
		addOrRemove, name, secCmdVarErr := message.getDualCommandParameters()
		if secCmdVarErr != nil {
			bot.ReplyToMessage(message, "allowlist <add|remove> <name>")
			return
		}

		if addOrRemove == "add" {
			err := bot.allowlist.Add(name)
			if err == nil {
				bot.ReplyToMessage(message, fmt.Sprintf("added %s to the allowlist", name))
			} else {
				bot.ReplyToMessage(message, fmt.Sprintf("error: %v", err))
			}
		} else if addOrRemove == "remove" {
			err := bot.allowlist.Remove(name)
			if err == nil {
				bot.ReplyToMessage(message, fmt.Sprintf("removed %s from the allowlist", name))
			} else {
				bot.ReplyToMessage(message, fmt.Sprintf("error: %v", err))
			}
		} else {
			bot.ReplyToMessage(message, "allowlist <add|remove> <name>")
			return
		}
	},
}

var volCommand = Command{
	Name:    "vol",
	Aliases: []string{"v"},
	Function: func(bot *MusicBot, message Message) {
		volumeString, commandVariableErr := message.getCommandParameter()

		if commandVariableErr != nil {
			volume, err := bot.musicPlayer.GetVolume()

			if err != nil {
				bot.ReplyToMessage(message, fmt.Sprintf("unable to get volume: %v", err))
				return
			}

			bot.ReplyToMessage(message, fmt.Sprintf("Current volume: %d", volume))
			return
		}

		// init vars here so we can use them after the switch statement
		var volume int
		var err error

		switch volumeString {
		case "+":
			{
				volume, err = bot.musicPlayer.IncreaseVolume(5)
				if err != nil {
					bot.ReplyToMessage(message, fmt.Sprintf("unable to increase volume: %s", err))
					return
				}
			}
		case "++":
			{
				volume, err = bot.musicPlayer.IncreaseVolume(10)
				if err != nil {
					bot.ReplyToMessage(message, fmt.Sprintf("unable to increase volume: %s", err))
					return
				}
			}
		case "+++":
			{
				volume, err = bot.musicPlayer.IncreaseVolume(20)
				if err != nil {
					bot.ReplyToMessage(message, fmt.Sprintf("unable to increase volume: %s", err))
					return
				}
			}
		case "-":
			{
				volume, err = bot.musicPlayer.DecreaseVolume(5)
				if err != nil {
					bot.ReplyToMessage(message, fmt.Sprintf("unable to decrease volume: %s", err))
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
		case "---":
			{
				volume, err = bot.musicPlayer.DecreaseVolume(20)
				if err != nil {
					bot.ReplyToMessage(message, fmt.Sprintf("unable to decrease volume: %s", err))
					return
				}
			}
		default:
			{
				volume, err = strconv.Atoi(strings.TrimSpace(volumeString))

				if err != nil {
					bot.ReplyToMessage(message, fmt.Sprintf("%s is not a valid number", volumeString))
					return
				}

				if volume >= 0 && volume <= 100 {
					_ = bot.musicPlayer.SetVolume(volume)
				} else {
					bot.ReplyToMessage(message, fmt.Sprintf("%s is not a valid volume", volumeString))
					return
				}
			}
		}

		if message.IsPrivate {
			bot.BroadcastMessage(fmt.Sprintf("Volume set to %d by %s", volume, message.Sender.Name))
		}

		bot.ReplyToMessage(message, fmt.Sprintf("Volume set to %d", volume))
	},
}

var addPlaylistCommand = Command{
	Name:    "playlist-add",
	Aliases: []string{"pa"},
	Function: func(bot *MusicBot, message Message) {
		parameter, commandVariableErr := message.getCommandParameter()

		if commandVariableErr != nil {
			bot.ReplyToMessage(message, "No URL provided")
			return
		}

		playlist, err := bot.musicPlayer.AddPlaylist(parameter)
		if err != nil {
			bot.ReplyToMessage(message, fmt.Sprintf("error: %v", err))
			return
		}

		bot.ReplyToMessage(message, fmt.Sprintf("Started playing Playlist '%s' with %d songs", playlist.Title, playlist.Length()))

	},
}

var aboutCommand = Command{
	Name: "about",
	Function: func(bot *MusicBot, message Message) {
		var GoVersion, Version, BuildDate string

		info, ok := debug.ReadBuildInfo()
		if ok {
			GoVersion = info.GoVersion

			for _, setting := range info.Settings {
				switch setting.Key {
				case "vcs.revision":
					Version = setting.Value
				case "vcs.time":
					BuildDate = setting.Value
				}
			}
		}

		bot.ReplyToMessage(message, "go-MusicBot by Sven Wiltink: https://github.com/svenwiltink/go-MusicBot")
		bot.ReplyToMessage(message, fmt.Sprintf("Go: %s", GoVersion))
		bot.ReplyToMessage(message, fmt.Sprintf("Version: %s", Version))
		bot.ReplyToMessage(message, fmt.Sprintf("Build date: %s", BuildDate))
	},
}
