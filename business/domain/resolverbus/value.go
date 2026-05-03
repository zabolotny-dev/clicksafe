package resolverbus

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type optionalValue interface {
	Valid() bool
}

func walkSegment(value reflect.Value, segment string) (reflect.Value, bool) {
	current, isMissing := dereferenceValue(value)
	if isMissing {
		return reflect.Value{}, true
	}

	switch current.Kind() {
	case reflect.Struct:
		field := current.FieldByName(segment)
		if !field.IsValid() || !field.CanInterface() {
			return reflect.Value{}, true
		}

		return field, false

	case reflect.Map:
		if current.Type().Key().Kind() != reflect.String {
			return reflect.Value{}, true
		}

		item := current.MapIndex(reflect.ValueOf(segment))
		if !item.IsValid() {
			return reflect.Value{}, true
		}

		return item, false

	default:
		return reflect.Value{}, true
	}
}

func stringifyValue(value reflect.Value) (string, bool, error) {
	current, isMissing := dereferenceValue(value)
	if isMissing {
		return "", true, nil
	}

	if isMissingOptionalValue(current) {
		return "", true, nil
	}

	if !current.CanInterface() {
		return "", true, nil
	}

	if stringer, ok := current.Interface().(fmt.Stringer); ok {
		return stringer.String(), false, nil
	}

	switch current.Kind() {
	case reflect.Map, reflect.Struct, reflect.Slice, reflect.Array:
		data, err := json.Marshal(current.Interface())
		if err != nil {
			return "", false, fmt.Errorf("stringify %s: %w", current.Type(), err)
		}

		return string(data), false, nil

	default:
		return fmt.Sprint(current.Interface()), false, nil
	}
}

func dereferenceValue(value reflect.Value) (reflect.Value, bool) {
	current := value

	for current.IsValid() {
		switch current.Kind() {
		case reflect.Interface, reflect.Pointer:
			if current.IsNil() {
				return reflect.Value{}, true
			}

			current = current.Elem()

		default:
			return current, false
		}
	}

	return reflect.Value{}, true
}

func isMissingOptionalValue(value reflect.Value) bool {
	if !value.IsValid() || value.Kind() != reflect.Struct {
		return false
	}

	if value.CanInterface() {
		if optional, ok := value.Interface().(optionalValue); ok {
			return !optional.Valid()
		}
	}

	validField := value.FieldByName("valid")
	if validField.IsValid() && validField.Kind() == reflect.Bool {
		return !validField.Bool()
	}

	return false
}
