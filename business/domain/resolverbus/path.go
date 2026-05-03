package resolverbus

import (
	"fmt"
	"strings"
)

const (
	pathSeparator = "."

	rootEmployee     = "Employee"
	rootDepartment   = "Department"
	rootOrganization = "Organization"
)

func splitPath(path string) ([]string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("%w: path is empty", ErrUnsupportedPath)
	}

	segments := strings.Split(path, pathSeparator)
	if len(segments) < 2 {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedPath, path)
	}

	for _, segment := range segments {
		if segment == "" {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedPath, path)
		}
	}

	return segments, nil
}
