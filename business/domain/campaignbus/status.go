package campaignbus

import "fmt"

// =============================================================================
// Campaign Status

type CampaignStatus struct {
	value string
}

func (s CampaignStatus) String() string {
	return s.value
}

var cmpStatuses = make(map[string]CampaignStatus)

var (
	Draft     = newCmpStatus("DRAFT")
	Active    = newCmpStatus("ACTIVE")
	Paused    = newCmpStatus("PAUSED")
	Completed = newCmpStatus("COMPLETED")
	Canceled  = newCmpStatus("CANCELED")
)

func newCmpStatus(s string) CampaignStatus {
	v := CampaignStatus{s}
	cmpStatuses[s] = v
	return v
}

func ParseCampaignStatus(value string) (CampaignStatus, error) {
	s, ok := cmpStatuses[value]
	if !ok {
		return CampaignStatus{}, fmt.Errorf("invalid campaign status: '%s'", value)
	}
	return s, nil
}

// =============================================================================
// Target Status

type TargetStatus struct {
	value string
}

func (s TargetStatus) String() string {
	return s.value
}

var targetStatuses = make(map[string]TargetStatus)

var (
	Pending   = newTargetStatus("PENDING")
	Sent      = newTargetStatus("SENT")
	Failed    = newTargetStatus("FAILED")
	Opened    = newTargetStatus("OPENED")
	Clicked   = newTargetStatus("CLICKED")
	Submitted = newTargetStatus("SUBMITTED")
)

func newTargetStatus(s string) TargetStatus {
	v := TargetStatus{s}
	targetStatuses[s] = v
	return v
}

func ParseTargetStatus(value string) (TargetStatus, error) {
	s, ok := targetStatuses[value]
	if !ok {
		return TargetStatus{}, fmt.Errorf("invalid target status: '%s'", value)
	}
	return s, nil
}
