package event

import "fmt"

var (
	MessageSent    = newEvent("MESSAGE_SENT")
	DeliveryFailed = newEvent("DELIVERY_FAILED")
	EMAIL_OPENED   = newEvent("EMAIL_OPENED")
	LINK_OPENED    = newEvent("LINK_OPENED")
	DATA_SENT      = newEvent("DATA_SENT")
)

var events = make(map[string]EventType)

type EventType struct {
	value string
}

func newEvent(event string) EventType {
	e := EventType{event}
	events[event] = e
	return e
}

func Parse(value string) (EventType, error) {
	e, ok := events[value]
	if !ok {
		return EventType{}, fmt.Errorf("invalid event type: '%s'", value)
	}

	return e, nil
}

func (e EventType) String() string {
	return e.value
}
