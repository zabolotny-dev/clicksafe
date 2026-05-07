package targetdb

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

type dbFilter struct {
	ID          *uuid.UUID
	CampaignID  *uuid.UUID
	EmployeeID  *uuid.UUID
	Status      pgtype.Text
	HasSchedule pgtype.Bool
}

func toDBFilter(filter campaignbus.TargetQueryFilter) dbFilter {
	var status pgtype.Text
	if filter.Status != nil {
		status = pgtype.Text{String: filter.Status.String(), Valid: true}
	}

	var hasSchedule pgtype.Bool
	if filter.HasSchedule != nil {
		hasSchedule = pgtype.Bool{Bool: *filter.HasSchedule, Valid: true}
	}

	return dbFilter{ID: filter.ID,
		CampaignID:  filter.CampaignID,
		EmployeeID:  filter.EmployeeID,
		Status:      status,
		HasSchedule: hasSchedule}
}
