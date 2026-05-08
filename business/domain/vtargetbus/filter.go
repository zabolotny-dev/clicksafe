package vtargetbus

import "github.com/google/uuid"

type Filter struct {
	ID         *uuid.UUID
	CampaignID *uuid.UUID
	EmployeeID *uuid.UUID
	Status     *string
	FullName   *string
}
