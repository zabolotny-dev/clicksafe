package campaignbus

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/types/date"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type Campaign struct {
	ID         uuid.UUID
	MessageID  *uuid.UUID
	Label      label.Label
	Status     Status
	DateRange  date.Null
	Attributes map[string]string
}

type NewCampaign struct {
	MessageID  *uuid.UUID
	Label      label.Label
	DateRange  date.Null
	Attributes map[string]string
}

type UpdateCampaign struct {
	MessageID  *uuid.UUID
	Label      *label.Label
	DateRange  *date.Null
	Attributes *map[string]string
}
