package bot

import (
	"strings"
)

type Sender struct {
	Name     string
	NickName string
}

type Message struct {
	Message   string
	Sender    Sender
	Target    string
	IsPrivate bool
}

func (message Message) getCommandWord() string {
	words := strings.SplitN(message.Message, " ", 2)
	return strings.TrimSpace(words[0])
}

func (message Message) getCommandParameter() (string, error) {
	parameters := strings.SplitN(message.Message, " ", 2)
	if len(parameters) <= 1 {
		return "", errVariableNotFound
	}

	parameter := strings.TrimSpace(parameters[1])

	if parameter == "" {
		return "", errVariableNotFound
	}

	return parameter, nil
}

func (message Message) getDualCommandParameters() (string, string, error) {
	parameters := strings.SplitN(message.Message, " ", 3)
	if len(parameters) <= 2 {
		return "", "", errVariableNotFound
	}

	parameter1 := strings.TrimSpace(parameters[1])
	parameter2 := strings.TrimSpace(parameters[2])

	if parameter1 == "" || parameter2 == "" {
		return "", "", errVariableNotFound
	}

	return parameter1, parameter2, nil
}
