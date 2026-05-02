package departmentdb

import (
	"encoding/json"

	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus/stores/departmentdb/sqlc"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
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

	dLabel, err := label.Parse(d.Name)
	if err != nil {
		return departmentbus.Department{}, err
	}

	return departmentbus.Department{
		ID:         d.ID,
		Name:       dLabel,
		Attributes: attributes,
	}, nil
}

func toBusDepartments(departments []sqlc.Department) ([]departmentbus.Department, error) {
	busDepartments := make([]departmentbus.Department, len(departments))

	for i, e := range departments {
		var err error
		busDepartments[i], err = toBusDepartment(e)
		if err != nil {
			return nil, err
		}
	}

	return busDepartments, nil
}
