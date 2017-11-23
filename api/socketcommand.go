package api

type SocketCommand struct {
	authRequired bool
	callback     func(*Command)

	cws *ControlWebsocket
}

func NewSocketCommand(cws *ControlWebsocket, authRequired bool, callback func(*Command)) (sc *SocketCommand) {
	return &SocketCommand{
		authRequired: authRequired,
		callback:     callback,
		cws:          cws,
	}
}

func (sc SocketCommand) execute(command *Command) {
	if sc.authRequired && !sc.cws.authenticated {
		return
	}

	sc.callback(command)
}
