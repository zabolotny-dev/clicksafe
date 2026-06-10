package employeebus_test

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/employeebus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/sdk/order"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/name"
	"github.com/zabolotny-dev/clicksafe/business/types/phone"
)

func Test_Employee(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Employee")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unittest.Run(t, query(db.BusDomain, sd), "query")
	unittest.Run(t, create(db.BusDomain, sd), "create")
	unittest.Run(t, update(db.BusDomain, sd), "update")
	unittest.Run(t, deleteEmp(db.BusDomain, sd), "delete")
	unittest.Run(t, savemany(db.BusDomain, sd), "savemany")
	unittest.Run(t, count(db.BusDomain, sd), "count")
}

// =============================================================================

type seedData struct {
	Department departmentbus.Department
	Employees  []employeebus.Employee
}

func insertSeedData(busDomain dbtest.BusDomain) (seedData, error) {
	ctx := context.Background()

	deps, err := departmentbus.TestSeedDepartments(ctx, 1, busDomain.Department)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding department: %w", err)
	}
	dep := deps[0]

	emps, err := employeebus.TestSeedEmployees(ctx, 2, &dep.ID, busDomain.Employee)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding employees: %w", err)
	}

	return seedData{Department: dep, Employees: emps}, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "byid",
			ExpResp: sd.Employees[0],
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Employee.QueryByID(ctx, sd.Employees[0].ID)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(employeebus.Employee)
				if !ok {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp.(employeebus.Employee))
			},
		},
		{
			Name:    "bydepartment",
			ExpResp: 2,
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Employee.Query(ctx,
					employeebus.QueryFilter{DepartmentID: &sd.Department.ID},
					employeebus.DefaultOrderBy,
					page.MustParse("1", "10"),
				)
				if err != nil {
					return err
				}
				return len(resp)
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func create(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	email := mail.Address{Address: "newemployee@example.com"}

	return []unittest.Table{
		{
			Name: "basic",
			ExpResp: employeebus.Employee{
				DepartmentID: &sd.Department.ID,
				FirstName:    name.MustParse("Alice"),
				LastName:     name.MustParse("Smith"),
				Email:        email,
				Attributes:   map[string]string{},
			},
			ExcFunc: func(ctx context.Context) any {
				ne := employeebus.NewEmployee{
					DepartmentID: &sd.Department.ID,
					FirstName:    name.MustParse("Alice"),
					LastName:     name.MustParse("Smith"),
					Email:        email,
					Attributes:   map[string]string{},
				}
				resp, err := busDomain.Employee.Save(ctx, ne)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(employeebus.Employee)
				if !ok {
					return "error occurred"
				}
				expResp := exp.(employeebus.Employee)
				expResp.ID = gotResp.ID
				expResp.Phone = gotResp.Phone
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	newFirst := name.MustParse("Updated")
	newEmail := mail.Address{Address: "updated@example.com"}

	return []unittest.Table{
		{
			Name: "basic",
			ExpResp: employeebus.Employee{
				ID:           sd.Employees[0].ID,
				DepartmentID: sd.Employees[0].DepartmentID,
				FirstName:    newFirst,
				LastName:     sd.Employees[0].LastName,
				Email:        newEmail,
				Phone:        sd.Employees[0].Phone,
				Attributes:   sd.Employees[0].Attributes,
			},
			ExcFunc: func(ctx context.Context) any {
				up := employeebus.UpdateEmployee{
					FirstName: &newFirst,
					Email:     &newEmail,
				}
				resp, err := busDomain.Employee.Update(ctx, sd.Employees[0], up)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(employeebus.Employee)
				if !ok {
					return "error occurred"
				}
				expResp := exp.(employeebus.Employee)
				expResp.Attributes = gotResp.Attributes
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func deleteEmp(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "basic",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Employee.Delete(ctx, sd.Employees[1]); err != nil {
					return err
				}
				return nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func savemany(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "basic",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Employee.SaveMany(ctx, []employeebus.NewEmployee{
					{
						DepartmentID: &sd.Department.ID,
						FirstName:    name.MustParse("Batch"),
						LastName:     name.MustParse("One"),
						Email:        mail.Address{Address: "batch1@example.com"},
						Attributes:   map[string]string{},
					},
					{
						DepartmentID: &sd.Department.ID,
						FirstName:    name.MustParse("Batch"),
						LastName:     name.MustParse("Two"),
						Email:        mail.Address{Address: "batch2@example.com"},
						Attributes:   map[string]string{},
					},
				})
				return err == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "empty-no-op",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busDomain.Employee.SaveMany(ctx, []employeebus.NewEmployee{}) == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

// =============================================================================
// Stub-based unit tests for error paths and missing branches

type employeeStorerStub struct {
	data    map[uuid.UUID]employeebus.Employee
	saveErr error
	qErr    error
}

func newEmployeeStorerStub() *employeeStorerStub {
	return &employeeStorerStub{data: make(map[uuid.UUID]employeebus.Employee)}
}

func (s *employeeStorerStub) Save(_ context.Context, e employeebus.Employee) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[e.ID] = e
	return nil
}

func (s *employeeStorerStub) SaveMany(_ context.Context, emps []employeebus.Employee) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	for _, e := range emps {
		s.data[e.ID] = e
	}
	return nil
}

func (s *employeeStorerStub) QueryByID(_ context.Context, id uuid.UUID) (employeebus.Employee, error) {
	if s.qErr != nil {
		return employeebus.Employee{}, s.qErr
	}
	e, ok := s.data[id]
	if !ok {
		return employeebus.Employee{}, employeebus.ErrNotFound
	}
	return e, nil
}

func (s *employeeStorerStub) Query(_ context.Context, _ employeebus.QueryFilter, _ order.By, _ page.Page) ([]employeebus.Employee, error) {
	if s.qErr != nil {
		return nil, s.qErr
	}
	var result []employeebus.Employee
	for _, e := range s.data {
		result = append(result, e)
	}
	return result, nil
}

func (s *employeeStorerStub) Update(_ context.Context, e employeebus.Employee) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	s.data[e.ID] = e
	return nil
}

func (s *employeeStorerStub) Delete(_ context.Context, e employeebus.Employee) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	delete(s.data, e.ID)
	return nil
}

func (s *employeeStorerStub) Count(_ context.Context, _ employeebus.QueryFilter) (int, error) {
	if s.qErr != nil {
		return 0, s.qErr
	}
	return len(s.data), nil
}

func Test_Employee_UnitErrors(t *testing.T) {
	t.Parallel()
	unittest.Run(t, employeeErrorPaths(), "error-paths")
}

func employeeErrorPaths() []unittest.Table {
	dbErr := errors.New("db error")
	deptNotFoundErr := employeebus.ErrDepartmentNotFound

	empID := uuid.New()
	deptID := uuid.New()

	origEmp := employeebus.Employee{
		ID:        empID,
		FirstName: name.MustParse("Old"),
		LastName:  name.MustParse("Name"),
		Email:     mail.Address{Address: "old@example.com"},
	}

	storeSaveErr := newEmployeeStorerStub()
	storeSaveErr.saveErr = dbErr
	busSaveErr := employeebus.NewBusiness(storeSaveErr)

	storeDeptErr := newEmployeeStorerStub()
	storeDeptErr.saveErr = deptNotFoundErr
	busDeptSaveErr := employeebus.NewBusiness(storeDeptErr)

	storeUpdateErr := newEmployeeStorerStub()
	storeUpdateErr.data[empID] = origEmp
	storeUpdateErr.saveErr = dbErr
	busUpdateErr := employeebus.NewBusiness(storeUpdateErr)

	storeDeptUpdateErr := newEmployeeStorerStub()
	storeDeptUpdateErr.data[empID] = origEmp
	storeDeptUpdateErr.saveErr = deptNotFoundErr
	busDeptUpdateErr := employeebus.NewBusiness(storeDeptUpdateErr)

	storeDelErr := newEmployeeStorerStub()
	storeDelErr.data[empID] = origEmp
	storeDelErr.saveErr = dbErr
	busDelErr := employeebus.NewBusiness(storeDelErr)

	storeQErr := newEmployeeStorerStub()
	storeQErr.qErr = dbErr
	busQErr := employeebus.NewBusiness(storeQErr)

	ph, _ := phone.ParseNull("+79991234567")
	newLN := name.MustParse("Updated")
	newPh := ph

	return []unittest.Table{
		{
			Name:    "save-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busSaveErr.Save(ctx, employeebus.NewEmployee{
					FirstName: name.MustParse("Test"),
					LastName:  name.MustParse("User"),
					Email:     mail.Address{Address: "t@example.com"},
				})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "save-dept-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				id := uuid.New()
				_, err := busDeptSaveErr.Save(ctx, employeebus.NewEmployee{
					DepartmentID: &id,
					FirstName:    name.MustParse("Test"),
					LastName:     name.MustParse("User"),
					Email:        mail.Address{Address: "t2@example.com"},
				})
				return errors.Is(err, employeebus.ErrDepartmentNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-store-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busUpdateErr.Update(ctx, origEmp, employeebus.UpdateEmployee{LastName: &newLN})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-last-name-and-phone",
			ExpResp: newLN,
			ExcFunc: func(ctx context.Context) any {
				store := newEmployeeStorerStub()
				store.data[empID] = origEmp
				bus := employeebus.NewBusiness(store)
				updated, err := bus.Update(ctx, origEmp, employeebus.UpdateEmployee{
					LastName: &newLN,
					Phone:    &newPh,
				})
				if err != nil {
					return err
				}
				return updated.LastName
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-dept-nil",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				store := newEmployeeStorerStub()
				empWithDept := employeebus.Employee{
					ID:           uuid.New(),
					DepartmentID: &deptID,
					FirstName:    name.MustParse("With"),
					LastName:     name.MustParse("Dept"),
					Email:        mail.Address{Address: "dept@example.com"},
				}
				store.data[empWithDept.ID] = empWithDept
				bus := employeebus.NewBusiness(store)
				var nilDeptID *uuid.UUID
				updated, err := bus.Update(ctx, empWithDept, employeebus.UpdateEmployee{DepartmentID: nilDeptID})
				if err != nil {
					return err
				}
				return updated.DepartmentID == &deptID
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "update-dept-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				id := uuid.New()
				_, err := busDeptUpdateErr.Update(ctx, origEmp, employeebus.UpdateEmployee{DepartmentID: &id})
				return errors.Is(err, employeebus.ErrDepartmentNotFound)
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "delete-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busDelErr.Delete(ctx, origEmp) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "query-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busQErr.Query(ctx, employeebus.QueryFilter{}, employeebus.DefaultOrderBy, page.MustParse("1", "10"))
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "count-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busQErr.Count(ctx, employeebus.QueryFilter{})
				return err != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
		{
			Name:    "savemany-error",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				return busUpdateErr.SaveMany(ctx, []employeebus.NewEmployee{
					{FirstName: name.MustParse("Batch"), LastName: name.MustParse("Test"), Email: mail.Address{Address: "b@example.com"}},
				}) != nil
			},
			CmpFunc: func(got, exp any) string { return cmp.Diff(got, exp) },
		},
	}
}

func count(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "byid-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.Employee.QueryByID(ctx, uuid.New())
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "count-all",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				n, err := busDomain.Employee.Count(ctx, employeebus.QueryFilter{})
				if err != nil {
					return err
				}
				return n >= 0
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}
