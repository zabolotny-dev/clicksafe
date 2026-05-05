package campaignbus

import (
	"time"

	"github.com/google/uuid"
)

type QueryFilter struct {
	ID       *uuid.UUID
	Label    *string
	Status   *Status
	DateFrom *time.Time
	DateTo   *time.Time
}
