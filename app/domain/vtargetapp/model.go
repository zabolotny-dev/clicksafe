package vtargetapp

import (
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/vtargetbus"
)

type Target struct {
	ID          uuid.UUID  `json:"id"`
	Token       string     `json:"token"`
	CampaignID  uuid.UUID  `json:"campaign_id"`
	EmployeeID  uuid.UUID  `json:"employee_id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	CreatedAt   time.Time  `json:"created_at"`
	Events      []Event    `json:"events"`
}

type Event struct {
	Type       string      `json:"type"`
	IPAddress  *netip.Addr `json:"ip_address,omitempty"`
	UserAgent  string      `json:"user_agent,omitempty"`
	Referer    string      `json:"referer,omitempty"`
	OccurredAt time.Time   `json:"occurred_at"`
}

func toAppTarget(t vtargetbus.Target) Target {
	return Target{
		ID:          t.ID,
		Token:       t.Token,
		CampaignID:  t.CampaignID,
		EmployeeID:  t.EmployeeID,
		FirstName:   t.FirstName,
		LastName:    t.LastName,
		Status:      t.Status,
		ScheduledAt: t.ScheduledAt,
		CreatedAt:   t.CreatedAt,
		Events:      toAppEvents(t.Events),
	}
}

func toAppTargets(targets []vtargetbus.Target) []Target {
	result := make([]Target, len(targets))
	for i, target := range targets {
		result[i] = toAppTarget(target)
	}

	return result
}

func toAppEvents(events []vtargetbus.Event) []Event {
	result := make([]Event, len(events))
	for i, event := range events {
		var ipAddr *netip.Addr
		if event.IPAddress.IsValid() {
			addr := event.IPAddress
			ipAddr = &addr
		}

		result[i] = Event{
			Type:       event.Type,
			IPAddress:  ipAddr,
			UserAgent:  event.UserAgent,
			Referer:    event.Referer,
			OccurredAt: event.OccurredAt,
		}
	}

	return result
}
