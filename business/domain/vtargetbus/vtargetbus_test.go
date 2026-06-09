package vtargetbus

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
)

// =============================================================================
// Stubs

type vtargetStorerStub struct {
	targets  []Target
	count    int
	queryErr error
	countErr error
}

func (s *vtargetStorerStub) Query(_ context.Context, _ Filter, _ order.By, _ page.Page) ([]Target, error) {
	if s.queryErr != nil {
		return nil, s.queryErr
	}
	return s.targets, nil
}

func (s *vtargetStorerStub) Count(_ context.Context, _ Filter) (int, error) {
	if s.countErr != nil {
		return 0, s.countErr
	}
	return s.count, nil
}

// =============================================================================
// Query tests

func TestQuery_ReturnsTargets(t *testing.T) {
	t.Parallel()

	targets := []Target{
		{ID: uuid.New(), FirstName: "Ivan", LastName: "Ivanov", Status: "sent"},
		{ID: uuid.New(), FirstName: "Petr", LastName: "Petrov", Status: "pending"},
	}
	stub := &vtargetStorerStub{targets: targets, count: len(targets)}
	bus := NewBusiness(stub)

	got, err := bus.Query(context.Background(), Filter{}, DefaultOrderBy, page.Page{})
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if len(got) != len(targets) {
		t.Fatalf("Query returned %d targets, want %d", len(got), len(targets))
	}
}

func TestQuery_EmptyResult(t *testing.T) {
	t.Parallel()

	stub := &vtargetStorerStub{targets: nil, count: 0}
	bus := NewBusiness(stub)

	got, err := bus.Query(context.Background(), Filter{}, DefaultOrderBy, page.Page{})
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("Query returned %d targets, want 0", len(got))
	}
}

func TestQuery_PropagatesStorerError(t *testing.T) {
	t.Parallel()

	stub := &vtargetStorerStub{queryErr: errors.New("db failure")}
	bus := NewBusiness(stub)

	_, err := bus.Query(context.Background(), Filter{}, DefaultOrderBy, page.Page{})
	if err == nil {
		t.Fatal("Query expected error, got nil")
	}
}

func TestQuery_WithFilter_PassesThrough(t *testing.T) {
	t.Parallel()

	campaignID := uuid.New()
	employeeID := uuid.New()
	status := "sent"
	fullName := "Ivan"

	expected := []Target{
		{ID: uuid.New(), CampaignID: campaignID, EmployeeID: employeeID, Status: status},
	}
	stub := &vtargetStorerStub{targets: expected}
	bus := NewBusiness(stub)

	filter := Filter{
		CampaignID: &campaignID,
		EmployeeID: &employeeID,
		Status:     &status,
		FullName:   &fullName,
	}

	got, err := bus.Query(context.Background(), filter, DefaultOrderBy, page.Page{})
	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Query returned %d targets, want 1", len(got))
	}
	if got[0].CampaignID != campaignID {
		t.Errorf("CampaignID = %v, want %v", got[0].CampaignID, campaignID)
	}
}

// =============================================================================
// Count tests

func TestCount_ReturnsCount(t *testing.T) {
	t.Parallel()

	stub := &vtargetStorerStub{count: 42}
	bus := NewBusiness(stub)

	n, err := bus.Count(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if n != 42 {
		t.Errorf("Count = %d, want 42", n)
	}
}

func TestCount_ZeroWhenEmpty(t *testing.T) {
	t.Parallel()

	stub := &vtargetStorerStub{count: 0}
	bus := NewBusiness(stub)

	n, err := bus.Count(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if n != 0 {
		t.Errorf("Count = %d, want 0", n)
	}
}

func TestCount_PropagatesStorerError(t *testing.T) {
	t.Parallel()

	stub := &vtargetStorerStub{countErr: errors.New("db failure")}
	bus := NewBusiness(stub)

	_, err := bus.Count(context.Background(), Filter{})
	if err == nil {
		t.Fatal("Count expected error, got nil")
	}
}

func TestCount_WithCampaignIDFilter(t *testing.T) {
	t.Parallel()

	campaignID := uuid.New()
	stub := &vtargetStorerStub{count: 5}
	bus := NewBusiness(stub)

	n, err := bus.Count(context.Background(), Filter{CampaignID: &campaignID})
	if err != nil {
		t.Fatalf("Count returned error: %v", err)
	}
	if n != 5 {
		t.Errorf("Count = %d, want 5", n)
	}
}
