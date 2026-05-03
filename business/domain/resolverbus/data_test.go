package resolverbus

import (
	"errors"
	"reflect"
	"testing"
)

func TestInsertResolvedValue(t *testing.T) {
	t.Parallel()

	t.Run("EmployeeFirstName", func(t *testing.T) {
		data := map[string]any{}

		if err := insertResolvedValue(data, "Employee.FirstName", "Ivan"); err != nil {
			t.Fatalf("insertResolvedValue returned error: %v", err)
		}

		expected := map[string]any{
			"Employee": map[string]any{
				"FirstName": "Ivan",
			},
		}
		if !reflect.DeepEqual(data, expected) {
			t.Fatalf("data = %#v, want %#v", data, expected)
		}
	})

	t.Run("EmployeeEmailAddress", func(t *testing.T) {
		data := map[string]any{}

		if err := insertResolvedValue(data, "Employee.Email.Address", "ivan@example.com"); err != nil {
			t.Fatalf("insertResolvedValue returned error: %v", err)
		}

		expected := map[string]any{
			"Employee": map[string]any{
				"Email": map[string]any{
					"Address": "ivan@example.com",
				},
			},
		}
		if !reflect.DeepEqual(data, expected) {
			t.Fatalf("data = %#v, want %#v", data, expected)
		}
	})

	t.Run("SamePathSameValue", func(t *testing.T) {
		data := map[string]any{}

		if err := insertResolvedValue(data, "Employee.FirstName", "Ivan"); err != nil {
			t.Fatalf("first insert returned error: %v", err)
		}

		if err := insertResolvedValue(data, "Employee.FirstName", "Ivan"); err != nil {
			t.Fatalf("second insert returned error: %v", err)
		}
	})

	t.Run("SamePathDifferentValue", func(t *testing.T) {
		data := map[string]any{}

		if err := insertResolvedValue(data, "Employee.FirstName", "Ivan"); err != nil {
			t.Fatalf("first insert returned error: %v", err)
		}

		err := insertResolvedValue(data, "Employee.FirstName", "Petr")
		if !errors.Is(err, ErrUnsupportedPath) {
			t.Fatalf("insertResolvedValue error = %v, want %v", err, ErrUnsupportedPath)
		}
	})

	t.Run("ConflictLeafThenNested", func(t *testing.T) {
		data := map[string]any{}

		if err := insertResolvedValue(data, "Employee.Email", "ivan@example.com"); err != nil {
			t.Fatalf("first insert returned error: %v", err)
		}

		err := insertResolvedValue(data, "Employee.Email.Address", "ivan@example.com")
		if !errors.Is(err, ErrUnsupportedPath) {
			t.Fatalf("insertResolvedValue error = %v, want %v", err, ErrUnsupportedPath)
		}
	})

	t.Run("ConflictNestedThenLeaf", func(t *testing.T) {
		data := map[string]any{}

		if err := insertResolvedValue(data, "Employee.Email.Address", "ivan@example.com"); err != nil {
			t.Fatalf("first insert returned error: %v", err)
		}

		err := insertResolvedValue(data, "Employee.Email", "ivan@example.com")
		if !errors.Is(err, ErrUnsupportedPath) {
			t.Fatalf("insertResolvedValue error = %v, want %v", err, ErrUnsupportedPath)
		}
	})

	t.Run("InvalidPath", func(t *testing.T) {
		invalidPaths := []string{"", " ", "Employee", "Employee.", "Employee..FirstName", ".Employee.FirstName"}

		for _, path := range invalidPaths {
			err := insertResolvedValue(map[string]any{}, path, "value")
			if !errors.Is(err, ErrUnsupportedPath) {
				t.Fatalf("path %q error = %v, want %v", path, err, ErrUnsupportedPath)
			}
		}
	})
}
