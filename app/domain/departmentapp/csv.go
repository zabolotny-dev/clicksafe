package departmentapp

import (
	"github.com/zabolotny-dev/clicksafe/app/sdk/csv"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

type csvRowErrorValue struct {
	Row   int    `json:"row"`
	Field string `json:"field,omitempty"`
	Err   string `json:"error"`
}

func csvToBusNewDepartment(row csv.Row) (departmentbus.NewDepartment, []csvRowErrorValue) {
	var rowErrors []csvRowErrorValue

	lbl, err := label.Parse(row.Get("label"))
	if err != nil {
		rowErrors = append(rowErrors, csvRowErrorValue{
			Row:   row.Number(),
			Field: "label",
			Err:   err.Error(),
		})
	}

	if len(rowErrors) > 0 {
		return departmentbus.NewDepartment{}, rowErrors
	}

	return departmentbus.NewDepartment{
		Label:      lbl,
		Attributes: row.Prefixed("attr."),
	}, nil
}
