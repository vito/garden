package protocol

type RequestMessage struct {
	TTY bool
	Argv []string
	User string
}

type ResponseMessage struct {}
