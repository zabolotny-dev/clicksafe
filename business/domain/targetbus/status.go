package targetbus

import "fmt"

var (
	Pending   = newTarget("PENDING")
	Sent      = newTarget("SENT")
	Failed    = newTarget("FAILED")
	Opened    = newTarget("OPENED")
	Clicked   = newTarget("CLICKED")
	Submitted = newTarget("SUBMITTED")
)

var statuses = make(map[string]Status)

type Status struct {
	value string
}

func newTarget(status string) Status {
	e := Status{status}
	statuses[status] = e
	return e
}

func Parse(value string) (Status, error) {
	e, ok := statuses[value]
	if !ok {
		return Status{}, fmt.Errorf("%w: '%s'", ErrInvalidStatusValue, value)
	}

	return e, nil
}

func (s Status) String() string {
	return s.value
}
