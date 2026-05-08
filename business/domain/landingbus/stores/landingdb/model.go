package landingdb

import (
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/landingbus/stores/landingdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/file"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func toDBLanding(landing landingbus.Landing) sqlc.Landing {
	return sqlc.Landing{
		ID:           landing.ID,
		Label:        landing.Label.String(),
		ContentPath:  landing.ContentPath.ToSQLNullString(),
		RequiredVars: landing.RequiredVars,
	}
}

func toBusLanding(landing sqlc.Landing) (landingbus.Landing, error) {
	lbl, err := label.Parse(landing.Label)
	if err != nil {
		return landingbus.Landing{}, err
	}

	contentPath, err := file.ParseNull(landing.ContentPath.String)
	if err != nil {
		return landingbus.Landing{}, err
	}

	return landingbus.Landing{
		ID:           landing.ID,
		Label:        lbl,
		ContentPath:  contentPath,
		RequiredVars: landing.RequiredVars,
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
