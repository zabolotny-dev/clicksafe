package apitest

import (
	"github.com/zabolotny-dev/clicksafe/business/domain/adminbus"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

// SeedAdmin holds a seeded admin together with the session credentials needed
// to authenticate HTTP requests in tests.
type SeedAdmin struct {
	Admin     adminbus.Admin
	RawToken  string // plain-text session token — goes into the cookie
	CSRFToken string // CSRF token — goes into X-CSRF-Token header
}

// SeedData carries all seeded entities for a single test run.
type SeedData struct {
	Admins    []SeedAdmin
	Campaigns []campaignbus.Campaign
}
