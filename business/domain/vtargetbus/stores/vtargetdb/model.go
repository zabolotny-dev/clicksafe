package vtargetdb

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus/stores/vtargetdb/sqlc"
)

type dbFilter struct {
	ID         *uuid.UUID
	CampaignID *uuid.UUID
	EmployeeID *uuid.UUID
	Status     pgtype.Text
	FullName   pgtype.Text
}

func toDBFilter(filter vtargetbus.Filter) dbFilter {
	return dbFilter{
		ID:         filter.ID,
		CampaignID: filter.CampaignID,
		EmployeeID: filter.EmployeeID,
		Status:     toDBText(filter.Status),
		FullName:   toDBText(filter.FullName),
	}
}

func toBusTargets(rows []sqlc.QueryRow) []vtargetbus.Target {
	targets := make([]vtargetbus.Target, 0, len(rows))
	indexByID := make(map[uuid.UUID]int, len(rows))

	for _, row := range rows {
		idx, exists := indexByID[row.ID]
		if !exists {
			targets = append(targets, vtargetbus.Target{
				ID:          row.ID,
				Token:       row.Token,
				CampaignID:  row.CampaignID,
				EmployeeID:  row.EmployeeID,
				FirstName:   row.FirstName,
				LastName:    row.LastName,
				Status:      row.Status,
				ScheduledAt: toBusTimePtr(row.ScheduledAt),
				CreatedAt:   row.CreatedAt.Time.UTC(),
			})
			idx = len(targets) - 1
			indexByID[row.ID] = idx
		}

		event, ok := toBusEvent(row)
		if ok {
			targets[idx].Events = append(targets[idx].Events, event)
		}
	}

	return targets
}

func toBusTimePtr(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}

	t := value.Time.UTC()
	return &t
}

func toBusEvent(row sqlc.QueryRow) (vtargetbus.Event, bool) {
	if !row.EventType.Valid || !row.EventOccurredAt.Valid {
		return vtargetbus.Event{}, false
	}

	var ipAddr netip.Addr
	if row.EventIpAddress != nil {
		ipAddr = *row.EventIpAddress
	}

	return vtargetbus.Event{
		Type:       row.EventType.String,
		OccurredAt: row.EventOccurredAt.Time.UTC(),
		IPAddress:  ipAddr,
		UserAgent:  row.EventUserAgent.String,
		Referer:    row.EventReferer.String,
	}, true
}

func toDBText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{String: *value, Valid: true}
}
