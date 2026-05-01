package departmentapp

import "github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"

var orderByFields = map[string]string{
	"department_id": departmentbus.OrderByID,
	"name":          departmentbus.OrderByName,
}
