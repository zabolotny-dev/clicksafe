package maxaccountbus

type AccountStatus struct {
	value string
}

func (s AccountStatus) String() string {
	return s.value
}

var statuses = map[string]AccountStatus{}

var (
	StatusActive       = newStatus("ACTIVE")
	StatusConnected    = newStatus("CONNECTED")
	StatusDisconnected = newStatus("DISCONNECTED")
	StatusError        = newStatus("ERROR")
	StatusPendingLogin = newStatus("PENDING_LOGIN")
)

func newStatus(value string) AccountStatus {
	status := AccountStatus{value: value}
	statuses[value] = status
	return status
}

func ParseStatus(value string) (AccountStatus, error) {
	status, ok := statuses[value]
	if !ok {
		return AccountStatus{}, ErrInvalidStatus
	}
	return status, nil
}
