package attachmentdb

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/attachmentbus"
)

type dbFilter struct {
	Label pgtype.Text
	Type  pgtype.Text
}

func toDBFilter(filter attachmentbus.QueryFilter) dbFilter {
	var labelFilter pgtype.Text
	if filter.Label != nil {
		labelFilter = pgtype.Text{String: *filter.Label, Valid: true}
	}

	var typeFilter pgtype.Text
	if filter.Type != nil {
		typeFilter = pgtype.Text{String: filter.Type.String(), Valid: true}
	}

	return dbFilter{Label: labelFilter, Type: typeFilter}
}
