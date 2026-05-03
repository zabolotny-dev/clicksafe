package resolverbus

import (
	"fmt"
	"strings"
)

func insertResolvedValue(data map[string]any, path string, value string) error {
	if data == nil {
		return fmt.Errorf("%w: resolved data is nil", ErrUnsupportedPath)
	}

	segments, err := splitPath(path)
	if err != nil {
		return err
	}

	current := data
	for idx, segment := range segments {
		prefix := strings.Join(segments[:idx+1], pathSeparator)

		if idx == len(segments)-1 {
			existing, exists := current[segment]
			if !exists {
				current[segment] = value
				return nil
			}

			existingValue, ok := existing.(string)
			if ok && existingValue == value {
				return nil
			}

			return fmt.Errorf("%w: path conflict at %q", ErrUnsupportedPath, prefix)
		}

		next, exists := current[segment]
		if !exists {
			child := make(map[string]any)
			current[segment] = child
			current = child
			continue
		}

		child, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("%w: path conflict at %q", ErrUnsupportedPath, prefix)
		}

		current = child
	}

	return nil
}
