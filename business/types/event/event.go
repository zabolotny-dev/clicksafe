package event

import "fmt"

var (
	MessageSent    = newEvent("MESSAGE_SENT")
	DeliveryFailed = newEvent("DELIVERY_FAILED")
	EmailOpened    = newEvent("EMAIL_OPENED")
	LinkOpened     = newEvent("LINK_OPENED")
	DataSent       = newEvent("DATA_SENT")
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
