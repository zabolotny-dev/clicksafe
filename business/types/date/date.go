package date

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type Range struct {
	start time.Time
	end   time.Time
}

func Parse(start, end time.Time) (Range, error) {
	if start.IsZero() {
		return Range{}, fmt.Errorf("empty start date")
	}
	if end.IsZero() {
		return Range{}, fmt.Errorf("empty end date")
	}
	if !end.After(start) {
		return Range{}, fmt.Errorf("invalid date range: end must be after start")
	}

	return Range{
		start: start.UTC(),
		end:   end.UTC(),
	}, nil
}

func (r Range) IsActive(now time.Time) bool {
	now = now.UTC()
	return !now.Before(r.start) && now.Before(r.end)
}

func (r Range) Contains(t time.Time) bool {
	t = t.UTC()
	return !t.Before(r.start) && t.Before(r.end)
}

func (r Range) Start() time.Time {
	return r.start
}

func (r Range) End() time.Time {
	return r.end
}

// =============================================================================

type Null struct {
	value Range
	valid bool
}

func ParseNull(start, end time.Time) (Null, error) {
	if start.IsZero() && end.IsZero() {
		return Null{}, nil
	}
	rng, err := Parse(start, end)
	if err != nil {
		return Null{}, err
	}
	return Null{
		value: rng,
		valid: true,
	}, nil
}

func ParseSQLNullTimestamptz(start, end pgtype.Timestamptz) (Null, error) {
	var startTime time.Time
	if start.Valid {
		startTime = start.Time
	}

	var endTime time.Time
	if end.Valid {
		endTime = end.Time
	}

	return ParseNull(startTime, endTime)
}

func (n Null) Valid() bool {
	return n.valid
}

func (n Null) Range() Range {
	if !n.valid {
		return Range{}
	}
	return n.value
}

func (n Null) ToSQLNullTimestamptz() (pgtype.Timestamptz, pgtype.Timestamptz) {
	if !n.valid {
		return pgtype.Timestamptz{}, pgtype.Timestamptz{}
	}

	return pgtype.Timestamptz{
			Time:  n.value.Start(),
			Valid: true,
		},
		pgtype.Timestamptz{
			Time:  n.value.End(),
			Valid: true,
		}
}
