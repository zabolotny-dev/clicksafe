package eventbus

import (
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/event"
)

type Event struct {
	ID         uuid.UUID
	CampaignID uuid.UUID
	EmployeeID uuid.UUID
	Type       event.EventType
	OccurredAt time.Time
}
