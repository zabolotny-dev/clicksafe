package organizationdb

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus/stores/organizationdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/url"
)

func toDBOrganization(org organizationbus.Organization) (sqlc.Organization, error) {
	attributes, err := json.Marshal(org.Attributes)
	if err != nil {
		return sqlc.Organization{}, fmt.Errorf("db: %w", err)
	}

	return sqlc.Organization{
		ID:         org.ID,
		Name:       org.Name.String(),
		LogoUrl:    pgtype.Text{String: org.LogoURL.String(), Valid: !org.LogoURL.IsEmpty()},
		Attributes: attributes,
	}, nil
}

func toBusOrganization(org sqlc.Organization) (organizationbus.Organization, error) {
	var attributes map[string]string
	if len(org.Attributes) > 0 {
		if err := json.Unmarshal(org.Attributes, &attributes); err != nil {
			return organizationbus.Organization{}, fmt.Errorf("db: %w", err)
		}
	}

	var logoURL url.URL
	if org.LogoUrl.Valid {
		var err error
		logoURL, err = url.Parse(org.LogoUrl.String)
		if err != nil {
			return organizationbus.Organization{}, fmt.Errorf("db: %w", err)
		}
	}

	orgName, err := name.Parse(org.Name)
	if err != nil {
		return organizationbus.Organization{}, fmt.Errorf("db: %w", err)
	}

	return organizationbus.Organization{
		ID:         org.ID,
		Name:       orgName,
		LogoURL:    logoURL,
		Attributes: attributes,
	}, nil
}
