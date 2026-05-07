package targetdb

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus/stores/targetdb/sqlc"
)

func toDBTarget(t campaignbus.Target) sqlc.Target {
	var schedat pgtype.Timestamptz
	if t.ScheduledAt != nil {
		schedat = pgtype.Timestamptz{Time: *t.ScheduledAt, Valid: true}
	}

	return sqlc.Target{
		ID:          t.ID,
		Token:       t.Token,
		EmployeeID:  t.EmployeeID,
		CampaignID:  t.CampaignID,
		Status:      t.Status.String(),
		ScheduledAt: schedat,
		CreatedAt:   pgtype.Timestamptz{Time: t.CreatedAt, Valid: true},
	}
}

func toBusTarget(t sqlc.Target) (campaignbus.Target, error) {
	status, err := campaignbus.ParseTargetStatus(t.Status)
	if err != nil {
		return campaignbus.Target{}, err
	}

	var scheduledAt *time.Time
	if t.ScheduledAt.Valid {
		scheduledAt = &t.ScheduledAt.Time
	}

	return campaignbus.Target{
		ID:          t.ID,
		Token:       t.Token,
		EmployeeID:  t.EmployeeID,
		CampaignID:  t.CampaignID,
		Status:      status,
		ScheduledAt: scheduledAt,
		CreatedAt:   t.CreatedAt.Time,
	}, nil
}

func toBusTargets(targets []sqlc.Target) ([]campaignbus.Target, error) {
	busTargets := make([]campaignbus.Target, len(targets))
	for i, t := range targets {
		var err error
		busTargets[i], err = toBusTarget(t)
		if err != nil {
			return nil, err
		}
	}
	return busTargets, nil
}
