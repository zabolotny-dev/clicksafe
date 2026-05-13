package organizationdb

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/organizationbus/stores/organizationdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func toDBOrganization(org organizationbus.Organization) (sqlc.Organization, error) {
	attributes, err := json.Marshal(org.Attributes)
	if err != nil {
		return sqlc.Organization{}, err
	}

	var id *uuid.UUID
	if org.AttachmentID.Valid {
		id = &org.AttachmentID.UUID
	}

	return sqlc.Organization{
		ID:           org.ID,
		Label:        org.Label.String(),
		AttachmentID: id,
		Attributes:   attributes,
	}, nil
}

func toBusOrganization(org sqlc.Organization) (organizationbus.Organization, error) {
	var attributes map[string]string
	if len(org.Attributes) > 0 {
		if err := json.Unmarshal(org.Attributes, &attributes); err != nil {
			return organizationbus.Organization{}, err
		}
	}

	var attachmentID uuid.NullUUID
	if org.AttachmentID != nil {
		attachmentID = uuid.NullUUID{
			UUID:  *org.AttachmentID,
			Valid: true,
		}
	}

	orgLabel, err := label.Parse(org.Label)
	if err != nil {
		return organizationbus.Organization{}, err
	}

	return organizationbus.Organization{
		ID:           org.ID,
		Label:        orgLabel,
		AttachmentID: attachmentID,
		Attributes:   attributes,
	}, nil
}
