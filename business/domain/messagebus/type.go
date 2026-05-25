package messagebus

import "fmt"

type MessageType struct {
	value string
}

func (t MessageType) String() string {
	return t.value
}

var messageTypes = make(map[string]MessageType)

var (
	EmailMessage = newMessageType("EMAIL")
	MaxMessage   = newMessageType("MAX")
)

func newMessageType(value string) MessageType {
	t := MessageType{value: value}
	messageTypes[value] = t
	return t
}

func ParseMessageType(value string) (MessageType, error) {
	t, ok := messageTypes[value]
	if !ok {
		return MessageType{}, fmt.Errorf("invalid message type: '%s'", value)
	}
	return t, nil
}
