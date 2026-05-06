package targetbus

import (
	"time"

	"github.com/google/uuid"
)

type Target struct {
	ID          uuid.UUID
	Token       string
	EmployeeID  uuid.UUID
	CampaignID  uuid.UUID
	Status      Status
	ScheduledAt *time.Time
	CreatedAt   time.Time
}

type NewTarget struct {
	EmployeeID uuid.UUID
	CampaignID uuid.UUID
}
