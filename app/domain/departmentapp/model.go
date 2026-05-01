package departmentapp

import (
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/app/sdk/errs"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
)

type Department struct {
	ID         uuid.UUID         `json:"id"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
}

type NewDepartment struct {
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
}

type UpdateDepartment struct {
	Name       *string            `json:"name"`
	Attributes *map[string]string `json:"attributes"`
}

func toAppDepartment(d departmentbus.Department) Department {
	return Department{
		ID:         d.ID,
		Name:       d.Name.String(),
		Attributes: d.Attributes,
	}
}

func toBusNewDepartment(d NewDepartment) (departmentbus.NewDepartment, error) {
	var errors errs.FieldErrors

	nm, err := name.Parse(d.Name)
	if err != nil {
		errors.Add("name", err)
	}

	if len(errors) > 0 {
		return departmentbus.NewDepartment{}, errors.ToError()
	}

	return departmentbus.NewDepartment{
		Name:       nm,
		Attributes: d.Attributes,
	}, nil
}

func toBusUpdateDepartment(d UpdateDepartment) (departmentbus.UpdateDepartment, error) {
	var errors errs.FieldErrors

	var nam *name.Name
	if d.Name != nil {
		nm, err := name.Parse(*d.Name)
		if err != nil {
			errors.Add("name", err)
		}
		nam = &nm
	}

	if len(errors) > 0 {
		return departmentbus.UpdateDepartment{}, errors.ToError()
	}

	return departmentbus.UpdateDepartment{
		Name:       nam,
		Attributes: d.Attributes,
	}, nil
}

func toAppDepartments(deps []departmentbus.Department) []Department {
	items := make([]Department, len(deps))
	for i, d := range deps {
		items[i] = toAppDepartment(d)
	}
	return items
}
