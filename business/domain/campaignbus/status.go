package campaignbus

import "fmt"

var (
	Draft     = newEvent("DRAFT")
	Active    = newEvent("ACTIVE")
	Paused    = newEvent("PAUSED")
	Completed = newEvent("COMPLETED")
	Canceled  = newEvent("CANCELED")
)

var statuses = make(map[string]Status)

type Status struct {
	value string
}

func newEvent(status string) Status {
	e := Status{status}
	statuses[status] = e
	return e
}

func Parse(value string) (Status, error) {
	e, ok := statuses[value]
	if !ok {
		return Status{}, fmt.Errorf("invalid status: '%s'", value)
	}

	return e, nil
}

func (s Status) String() string {
	return s.value
}
