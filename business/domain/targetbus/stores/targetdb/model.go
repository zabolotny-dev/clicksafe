package targetdb

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/targetbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/targetbus/stores/targetdb/sqlc"
)

func toDBTarget(t targetbus.Target) sqlc.Target {
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

func toBusTarget(t sqlc.Target) (targetbus.Target, error) {
	status, err := targetbus.Parse(t.Status)
	if err != nil {
		return targetbus.Target{}, err
	}

	var scheduledAt *time.Time
	if t.ScheduledAt.Valid {
		scheduledAt = &t.ScheduledAt.Time
	}

	return targetbus.Target{
		ID:          t.ID,
		Token:       t.Token,
		EmployeeID:  t.EmployeeID,
		CampaignID:  t.CampaignID,
		Status:      status,
		ScheduledAt: scheduledAt,
		CreatedAt:   t.CreatedAt.Time,
	}, nil
}

func toBusTargets(targets []sqlc.Target) ([]targetbus.Target, error) {
	busTargets := make([]targetbus.Target, len(targets))
	for i, t := range targets {
		var err error
		busTargets[i], err = toBusTarget(t)
		if err != nil {
			return nil, err
		}
	}
	return busTargets, nil
}
