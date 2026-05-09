package vtargetbus

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type Target struct {
	ID          uuid.UUID
	Token       string
	CampaignID  uuid.UUID
	EmployeeID  uuid.UUID
	FirstName   string
	LastName    string
	Status      string
	ScheduledAt *time.Time
	CreatedAt   time.Time
	Events      []Event
}

type Event struct {
	Type       string
	OccurredAt time.Time
	IPAddress  netip.Addr
	UserAgent  string
	Referer    string
}
