package eventbus

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/event"
)

type Event struct {
	ID         uuid.UUID
	CampaignID uuid.UUID
	EmployeeID uuid.UUID
	Type       event.EventType
	IPAddress  netip.Addr
	UserAgent  string
	Referer    string
	OccurredAt time.Time
}

type NewEvent struct {
	CampaignID uuid.UUID
	EmployeeID uuid.UUID
	Type       event.EventType
	IPAddress  netip.Addr
	UserAgent  string
	Referer    string
}
