package eventbus

import (
	"context"
	"errors"
	"net/netip"
	"testing"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Stubs

type eventStorerStub struct {
	saved   Event
	saveErr error
}

func (s *eventStorerStub) Save(_ context.Context, e Event) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.saved = e
	return nil
}

// =============================================================================
// Publish tests

func TestPublish_PersistsEvent(t *testing.T) {
	t.Parallel()

	stub := &eventStorerStub{}
	bus := NewBusinnes(stub)

	campaignID := uuid.New()
	employeeID := uuid.New()
	ip := netip.MustParseAddr("1.2.3.4")

	err := bus.Publish(context.Background(), NewEvent{
		CampaignID: campaignID,
		EmployeeID: employeeID,
		Type:       EmailOpened,
		IPAddress:  ip,
		UserAgent:  "Mozilla/5.0",
		Referer:    "https://example.com",
	})
	if err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	if stub.saved.ID == (uuid.UUID{}) {
		t.Error("Publish must assign a non-zero UUID")
	}
	if stub.saved.CampaignID != campaignID {
		t.Errorf("CampaignID = %v, want %v", stub.saved.CampaignID, campaignID)
	}
	if stub.saved.EmployeeID != employeeID {
		t.Errorf("EmployeeID = %v, want %v", stub.saved.EmployeeID, employeeID)
	}
	if stub.saved.Type != EmailOpened {
		t.Errorf("Type = %v, want %v", stub.saved.Type, EmailOpened)
	}
	if stub.saved.IPAddress != ip {
		t.Errorf("IPAddress = %v, want %v", stub.saved.IPAddress, ip)
	}
	if stub.saved.UserAgent != "Mozilla/5.0" {
		t.Errorf("UserAgent = %q, want %q", stub.saved.UserAgent, "Mozilla/5.0")
	}
	if stub.saved.Referer != "https://example.com" {
		t.Errorf("Referer = %q, want %q", stub.saved.Referer, "https://example.com")
	}
}

func TestPublish_AssignsUniqueIDs(t *testing.T) {
	t.Parallel()

	bus := NewBusinnes(&eventStorerStub{})

	stub1 := &eventStorerStub{}
	stub2 := &eventStorerStub{}
	bus1 := NewBusinnes(stub1)
	bus2 := NewBusinnes(stub2)

	_ = bus.Publish(context.Background(), NewEvent{Type: MessageSent})
	_ = bus1.Publish(context.Background(), NewEvent{Type: MessageSent})
	_ = bus2.Publish(context.Background(), NewEvent{Type: MessageSent})

	if stub1.saved.ID == stub2.saved.ID {
		t.Error("Publish must assign unique IDs per event")
	}
}

func TestPublish_SetsOccurredAtUTC(t *testing.T) {
	t.Parallel()

	stub := &eventStorerStub{}
	bus := NewBusinnes(stub)

	before := time.Now().UTC()
	_ = bus.Publish(context.Background(), NewEvent{Type: LinkOpened})
	after := time.Now().UTC()

	if stub.saved.OccurredAt.IsZero() {
		t.Error("OccurredAt must not be zero")
	}
	if stub.saved.OccurredAt.Before(before) || stub.saved.OccurredAt.After(after) {
		t.Errorf("OccurredAt = %v is not between %v and %v", stub.saved.OccurredAt, before, after)
	}
	if stub.saved.OccurredAt.Location() != time.UTC {
		t.Errorf("OccurredAt location = %v, want UTC", stub.saved.OccurredAt.Location())
	}
}

func TestPublish_PropagatesStorerError(t *testing.T) {
	t.Parallel()

	stub := &eventStorerStub{saveErr: errors.New("db failure")}
	bus := NewBusinnes(stub)

	err := bus.Publish(context.Background(), NewEvent{Type: DeliveryFailed})
	if err == nil {
		t.Fatal("Publish expected error, got nil")
	}
}

func TestPublish_AllEventTypes(t *testing.T) {
	t.Parallel()

	allTypes := []EventType{
		MessageSent, DeliveryFailed, MessageRead,
		MessageReplied, EmailOpened, LinkOpened, DataSent,
	}

	for _, et := range allTypes {
		et := et
		t.Run(et.String(), func(t *testing.T) {
			t.Parallel()

			stub := &eventStorerStub{}
			bus := NewBusinnes(stub)

			err := bus.Publish(context.Background(), NewEvent{Type: et})
			if err != nil {
				t.Fatalf("Publish(%s) returned error: %v", et, err)
			}
			if stub.saved.Type != et {
				t.Errorf("Type = %v, want %v", stub.saved.Type, et)
			}
		})
	}
}

// =============================================================================
// Parse tests

func TestParse_KnownEventTypes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected EventType
	}{
		{"MESSAGE_SENT", MessageSent},
		{"DELIVERY_FAILED", DeliveryFailed},
		{"MESSAGE_READ", MessageRead},
		{"MESSAGE_REPLIED", MessageReplied},
		{"EMAIL_OPENED", EmailOpened},
		{"LINK_OPENED", LinkOpened},
		{"DATA_SENT", DataSent},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Errorf("Parse(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestParse_UnknownEventType_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := Parse("UNKNOWN_EVENT")
	if err == nil {
		t.Fatal("Parse expected error for unknown event type, got nil")
	}
}

func TestParse_EmptyString_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := Parse("")
	if err == nil {
		t.Fatal("Parse expected error for empty string, got nil")
	}
}

// =============================================================================
// EventType.String tests

func TestEventType_String(t *testing.T) {
	t.Parallel()

	cases := []struct {
		et       EventType
		expected string
	}{
		{MessageSent, "MESSAGE_SENT"},
		{DeliveryFailed, "DELIVERY_FAILED"},
		{MessageRead, "MESSAGE_READ"},
		{MessageReplied, "MESSAGE_REPLIED"},
		{EmailOpened, "EMAIL_OPENED"},
		{LinkOpened, "LINK_OPENED"},
		{DataSent, "DATA_SENT"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()

			if got := tc.et.String(); got != tc.expected {
				t.Errorf("String() = %q, want %q", got, tc.expected)
			}
		})
	}
}
