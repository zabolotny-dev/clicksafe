package departmentdb

import (
	"encoding/json"

	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus/stores/departmentdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
)

func toDBDepartment(d departmentbus.Department) (sqlc.Department, error) {
	attributes, err := json.Marshal(d.Attributes)
	if err != nil {
		return sqlc.Department{}, err
	}

	return sqlc.Department{
		ID:         d.ID,
		Name:       d.Name.String(),
		Attributes: attributes,
	}, nil
}

func toBusDepartment(d sqlc.Department) (departmentbus.Department, error) {
	var attributes map[string]string
	if len(d.Attributes) > 0 {
		if err := json.Unmarshal(d.Attributes, &attributes); err != nil {
			return departmentbus.Department{}, err
		}
	}

	dName, err := name.Parse(d.Name)
	if err != nil {
		return departmentbus.Department{}, err
	}

	return departmentbus.Department{
		ID:         d.ID,
		Name:       dName,
		Attributes: attributes,
	}, nil
}
