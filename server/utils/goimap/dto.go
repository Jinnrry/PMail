package goimap

type CommandResponseType uint8

const (
	SUCCESS CommandResponseType = 0
	BAD     CommandResponseType = 1
	NO      CommandResponseType = 2
)

type CommandResponse struct {
	Type    CommandResponseType
	Message string
	Data    []string
}
