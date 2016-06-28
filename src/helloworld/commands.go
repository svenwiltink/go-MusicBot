package main

import "fmt"

type ICommand interface {
	execute(paramters []string) bool
	getName() string
}

type Command struct {
	Name string
	Function func(parameters []string) bool
}

func(c *Command) execute(parameters []string) bool {
	return c.Function(parameters)
}

var HelpCommand = Command{
	Name:"Help",
	Function:
	func(parameters []string) bool {
		fmt.Print(parameters)
		return true
	},
}