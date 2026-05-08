package resolverbus

import (
	"github.com/google/uuid"
)

type Scope struct {
	TargetID uuid.UUID
}

type Result struct {
	Data    map[string]any
	Missing []string
}
