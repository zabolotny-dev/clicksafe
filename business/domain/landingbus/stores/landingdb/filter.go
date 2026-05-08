package landingdb

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
)

type dbFilter struct {
	Label pgtype.Text
}

func toDBFilter(filter landingbus.QueryFilter) dbFilter {
	var labelFilter pgtype.Text
	if filter.Label != nil {
		labelFilter = pgtype.Text{String: *filter.Label, Valid: true}
	}

	return dbFilter{Label: labelFilter}
}
