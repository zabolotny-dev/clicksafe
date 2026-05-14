package landingdb

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus/stores/landingdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func toDBLanding(landing landingbus.Landing) sqlc.Landing {
	return sqlc.Landing{
		ID:         landing.ID,
		Label:      landing.Label.String(),
		HtmlBodyID: toDBHtmlBodyID(landing.HtmlBodyID),
	}
}

func toBusLanding(landing sqlc.Landing) (landingbus.Landing, error) {
	lbl, err := label.Parse(landing.Label)
	if err != nil {
		return landingbus.Landing{}, err
	}

	return landingbus.Landing{
		ID:         landing.ID,
		Label:      lbl,
		HtmlBodyID: toBusHtmlBodyID(landing.HtmlBodyID),
	}, nil
}

func toBusLandings(landings []sqlc.Landing) ([]landingbus.Landing, error) {
	busLandings := make([]landingbus.Landing, len(landings))

	for i, landing := range landings {
		var err error
		busLandings[i], err = toBusLanding(landing)
		if err != nil {
			return nil, err
		}
	}

	return busLandings, nil
}

func toDBHtmlBodyID(id uuid.NullUUID) *uuid.UUID {
	if !id.Valid {
		return nil
	}

	htmlBodyID := id.UUID
	return &htmlBodyID
}

func toBusHtmlBodyID(id *uuid.UUID) uuid.NullUUID {
	if id == nil {
		return uuid.NullUUID{}
	}

	return uuid.NullUUID{
		UUID:  *id,
		Valid: true,
	}
}
