package campaignbus

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// =============================================================================
// Common Errors

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// =============================================================================
// Campaign Errors

type ErrUnscheduledTargets struct {
	TargetIDs []uuid.UUID
}

func (e *ErrUnscheduledTargets) Error() string {
	return fmt.Sprintf("targets without schedule: %d", len(e.TargetIDs))
}

type ErrTargetsMissingVars struct {
	Targets []TargetMissingVars
}

func (e *ErrTargetsMissingVars) Error() string {
	return fmt.Sprintf("targets have missing required vars: %d", len(e.Targets))
}

var (
	ErrCampaignNotFound    = errors.New("campaign not found")
	ErrUniqueLabel         = errors.New("campaign with this label already exists")
	ErrMessageNotFound     = errors.New("message not found")
	ErrMessageRequired     = errors.New("campaign message is required")
	ErrMessageHTMLRequired = errors.New("campaign message has no HTML content")
	ErrLandingRequired     = errors.New("campaign landing is required")
	ErrLandingHTMLRequired = errors.New("campaign landing has no HTML content")
	ErrDateRangeRequired   = errors.New("campaign date range is required")
	ErrDateRangeExpired    = errors.New("campaign date range is expired")
	ErrTargetsRequired     = errors.New("campaign requires at least one target")
	ErrCampaignLocked      = errors.New("campaign locked")
	ErrDomainRequired      = errors.New("campaign domain is required")
	ErrLandingNotFound     = errors.New("landing not found")
)

// =============================================================================
// Target Errors

var (
	ErrTargetNotFound        = errors.New("target not found")
	ErrEmployeeNotFound      = errors.New("employee not found")
	ErrInvalidScheduleWindow = errors.New("invalid campaign schedule window")
	ErrUniqueToken           = errors.New("unique token constraint violation")
	ErrTargetAlreadyExists   = errors.New("target already exists")
	ErrTargetLocked          = errors.New("target locked")
)
