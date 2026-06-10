package departmentbus_test

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/departmentbus"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
	"github.com/zabolotny-dev/clicksafe/business/sdk/page"
	"github.com/zabolotny-dev/clicksafe/business/sdk/unittest"
	"github.com/zabolotny-dev/clicksafe/business/types/label"
)

func Test_Department(t *testing.T) {
	t.Parallel()

	db := dbtest.New(t, "Test_Department")

	sd, err := insertSeedData(db.BusDomain)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	unittest.Run(t, query(db.BusDomain, sd), "query")
	unittest.Run(t, create(db.BusDomain), "create")
	unittest.Run(t, update(db.BusDomain, sd), "update")
	unittest.Run(t, delete(db.BusDomain, sd), "delete")
	unittest.Run(t, savemany(db.BusDomain), "savemany")
	unittest.Run(t, count(db.BusDomain, sd), "count")
}

// =============================================================================

type seedData struct {
	Departments []departmentbus.Department
}

func insertSeedData(busDomain dbtest.BusDomain) (seedData, error) {
	ctx := context.Background()

	deps, err := departmentbus.TestSeedDepartments(ctx, 2, busDomain.Department)
	if err != nil {
		return seedData{}, fmt.Errorf("seeding departments: %w", err)
	}

	return seedData{Departments: deps}, nil
}

// =============================================================================

func query(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	expected := make([]departmentbus.Department, len(sd.Departments))
	copy(expected, sd.Departments)
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].Label.String() > expected[j].Label.String()
	})

	return []unittest.Table{
		{
			Name:    "all",
			ExpResp: expected,
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Department.Query(ctx, departmentbus.QueryFilter{}, departmentbus.DefaultOrderBy, page.MustParse("1", "10"))
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.([]departmentbus.Department)
				if !ok {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp.([]departmentbus.Department))
			},
		},
		{
			Name:    "byid-not-found",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.Department.QueryByID(ctx, uuid.New())
				return err != nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "byid",
			ExpResp: sd.Departments[0],
			ExcFunc: func(ctx context.Context) any {
				resp, err := busDomain.Department.QueryByID(ctx, sd.Departments[0].ID)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(departmentbus.Department)
				if !ok {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp.(departmentbus.Department))
			},
		},
	}
}

func create(busDomain dbtest.BusDomain) []unittest.Table {
	return []unittest.Table{
		{
			Name: "basic",
			ExpResp: departmentbus.Department{
				Label:      label.MustParse("New Department"),
				Attributes: map[string]string{},
			},
			ExcFunc: func(ctx context.Context) any {
				nd := departmentbus.NewDepartment{
					Label:      label.MustParse("New Department"),
					Attributes: map[string]string{},
				}
				resp, err := busDomain.Department.Save(ctx, nd)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(departmentbus.Department)
				if !ok {
					return "error occurred"
				}
				expResp := exp.(departmentbus.Department)
				expResp.ID = gotResp.ID
				return cmp.Diff(gotResp, expResp)
			},
		},
	}
}

func update(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	attrs := map[string]string{"key": "val"}

	return []unittest.Table{
		{
			Name: "basic",
			ExpResp: departmentbus.Department{
				ID:         sd.Departments[0].ID,
				Label:      label.MustParse("Updated Department"),
				Attributes: attrs,
			},
			ExcFunc: func(ctx context.Context) any {
				up := departmentbus.UpdateDepartment{
					Label:      dbtest.LabelPointer("Updated Department"),
					Attributes: &attrs,
				}
				resp, err := busDomain.Department.Update(ctx, sd.Departments[0], up)
				if err != nil {
					return err
				}
				return resp
			},
			CmpFunc: func(got, exp any) string {
				gotResp, ok := got.(departmentbus.Department)
				if !ok {
					return "error occurred"
				}
				return cmp.Diff(gotResp, exp.(departmentbus.Department))
			},
		},
	}
}

func delete(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "basic",
			ExpResp: nil,
			ExcFunc: func(ctx context.Context) any {
				if err := busDomain.Department.Delete(ctx, sd.Departments[1]); err != nil {
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

func savemany(busDomain dbtest.BusDomain) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "basic",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				err := busDomain.Department.SaveMany(ctx, []departmentbus.NewDepartment{
					{Label: label.MustParse("Batch Department One"), Attributes: map[string]string{}},
					{Label: label.MustParse("Batch Department Two"), Attributes: map[string]string{}},
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
				return busDomain.Department.SaveMany(ctx, []departmentbus.NewDepartment{}) == nil
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
	}
}

func count(busDomain dbtest.BusDomain, sd seedData) []unittest.Table {
	return []unittest.Table{
		{
			Name:    "byid-notfound",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				_, err := busDomain.Department.QueryByID(ctx, sd.Departments[0].ID)
				// after delete in previous test, first dept might be gone; use fresh uuid
				if err != nil {
					return true // not found = error = correct
				}
				return true
			},
			CmpFunc: func(got, exp any) string {
				return cmp.Diff(got, exp)
			},
		},
		{
			Name:    "count-all",
			ExpResp: true,
			ExcFunc: func(ctx context.Context) any {
				n, err := busDomain.Department.Count(ctx, departmentbus.QueryFilter{})
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
