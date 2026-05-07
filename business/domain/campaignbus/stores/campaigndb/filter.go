package campaigndb

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

type dbFilter struct {
	Label    pgtype.Text
	Status   pgtype.Text
	DateFrom pgtype.Timestamptz
	DateTo   pgtype.Timestamptz
}

func toDBFilter(filter campaignbus.CampaignQueryFilter) dbFilter {
	var labelFilter pgtype.Text
	if filter.Label != nil {
		labelFilter = pgtype.Text{String: *filter.Label, Valid: true}
	}

	var statusFilter pgtype.Text
	if filter.Status != nil {
		statusFilter = pgtype.Text{String: filter.Status.String(), Valid: true}
	}

	var dateFromFilter pgtype.Timestamptz
	if filter.DateFrom != nil {
		dateFromFilter = pgtype.Timestamptz{Time: filter.DateFrom.UTC(), Valid: true}
	}

	var dateToFilter pgtype.Timestamptz
	if filter.DateTo != nil {
		dateToFilter = pgtype.Timestamptz{Time: filter.DateTo.UTC(), Valid: true}
	}

	return dbFilter{Label: labelFilter, Status: statusFilter, DateFrom: dateFromFilter, DateTo: dateToFilter}
}
