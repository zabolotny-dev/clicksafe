package messagedb

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/messagebus"
)

type dbFilter struct {
	Type      pgtype.Text
	Label     pgtype.Text
	FromEmail pgtype.Text
	FromName  pgtype.Text
	Subject   pgtype.Text
}

func toDBFilter(filter messagebus.QueryFilter) dbFilter {
	var typeFilter pgtype.Text
	if filter.Type != nil {
		typeFilter = pgtype.Text{String: filter.Type.String(), Valid: true}
	}

	var labelFilter pgtype.Text
	if filter.Label != nil {
		labelFilter = pgtype.Text{String: *filter.Label, Valid: true}
	}

	var fromEmailFilter pgtype.Text
	if filter.FromEmail != nil {
		fromEmailFilter = pgtype.Text{String: *filter.FromEmail, Valid: true}
	}

	var fromNameFilter pgtype.Text
	if filter.FromName != nil {
		fromNameFilter = pgtype.Text{String: *filter.FromName, Valid: true}
	}

	var subjectFilter pgtype.Text
	if filter.Subject != nil {
		subjectFilter = pgtype.Text{String: *filter.Subject, Valid: true}
	}

	return dbFilter{Type: typeFilter, Label: labelFilter, FromEmail: fromEmailFilter, FromName: fromNameFilter, Subject: subjectFilter}
}
