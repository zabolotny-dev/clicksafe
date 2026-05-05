package date

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestNullToSQLNullTimestamptz(t *testing.T) {
	start := time.Date(2026, 5, 4, 10, 0, 0, 0, time.FixedZone("MSK", 3*60*60))
	end := start.Add(time.Hour)

	rng, err := ParseNull(start, end)
	if err != nil {
		t.Fatalf("ParseNull returned error: %v", err)
	}

	dateFrom, dateTo := rng.ToSQLNullTimestamptz()
	if !dateFrom.Valid {
		t.Fatalf("expected dateFrom to be valid")
	}
	if !dateTo.Valid {
		t.Fatalf("expected dateTo to be valid")
	}
	if !dateFrom.Time.Equal(start.UTC()) {
		t.Fatalf("expected dateFrom %v, got %v", start.UTC(), dateFrom.Time)
	}
	if !dateTo.Time.Equal(end.UTC()) {
		t.Fatalf("expected dateTo %v, got %v", end.UTC(), dateTo.Time)
	}
}

func TestNullToSQLNullTimestamptzAllowsEmptyRange(t *testing.T) {
	dateFrom, dateTo := Null{}.ToSQLNullTimestamptz()

	if dateFrom.Valid {
		t.Fatalf("expected dateFrom to be invalid")
	}
	if dateTo.Valid {
		t.Fatalf("expected dateTo to be invalid")
	}
}

func TestParseSQLNullTimestamptz(t *testing.T) {
	start := time.Date(2026, 5, 4, 7, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	rng, err := ParseSQLNullTimestamptz(
		pgtype.Timestamptz{Time: start, Valid: true},
		pgtype.Timestamptz{Time: end, Valid: true},
	)
	if err != nil {
		t.Fatalf("ParseSQLNullTimestamptz returned error: %v", err)
	}

	if !rng.Valid() {
		t.Fatalf("expected range to be valid")
	}
	if !rng.Range().Start().Equal(start) {
		t.Fatalf("expected start %v, got %v", start, rng.Range().Start())
	}
	if !rng.Range().End().Equal(end) {
		t.Fatalf("expected end %v, got %v", end, rng.Range().End())
	}
}

func TestParseSQLNullTimestamptzAllowsEmptyRange(t *testing.T) {
	rng, err := ParseSQLNullTimestamptz(pgtype.Timestamptz{}, pgtype.Timestamptz{})
	if err != nil {
		t.Fatalf("ParseSQLNullTimestamptz returned error: %v", err)
	}
	if rng.Valid() {
		t.Fatalf("expected range to be invalid")
	}
}

func TestParseSQLNullTimestamptzRejectsPartialRange(t *testing.T) {
	_, err := ParseSQLNullTimestamptz(
		pgtype.Timestamptz{Time: time.Now(), Valid: true},
		pgtype.Timestamptz{},
	)
	if err == nil {
		t.Fatalf("expected error")
	}
}
