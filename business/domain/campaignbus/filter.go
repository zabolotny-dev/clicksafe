package campaignbus

import (
	"time"

	"github.com/google/uuid"
)

type CampaignQueryFilter struct {
	ID       *uuid.UUID
	Label    *string
	Status   *CampaignStatus
	DateFrom *time.Time
	DateTo   *time.Time
}

type TargetQueryFilter struct {
	ID          *uuid.UUID
	CampaignID  *uuid.UUID
	EmployeeID  *uuid.UUID
	Status      *TargetStatus
	HasSchedule *bool
}
