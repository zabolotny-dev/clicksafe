package eventbus

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID         uuid.UUID
	CampaignID uuid.UUID
	EmployeeID uuid.UUID
	Type       EventType
	IPAddress  netip.Addr
	UserAgent  string
	Referer    string
	OccurredAt time.Time
}

type NewEvent struct {
	CampaignID uuid.UUID
	EmployeeID uuid.UUID
	Type       EventType
	IPAddress  netip.Addr
	UserAgent  string
	Referer    string
}
