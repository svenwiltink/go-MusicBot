package main

import (
	"github.com/thoj/go-ircevent"
)

type ICommand interface {
	execute(paramters []string) bool
	getName() string
}

type Command struct {
	Name string
	Function func(event *irc.Event, parameters []string) bool
}

func(c *Command) execute(event *irc.Event, parameters []string) bool {
	return c.Function(event, parameters)
}

var HelpCommand = Command {
	Name:"Help",
	Function:
	func(event *irc.Event, parameters []string) bool {
		event.Connection.Privmsg("#fuckit", event.Message())
		return true
	},
}