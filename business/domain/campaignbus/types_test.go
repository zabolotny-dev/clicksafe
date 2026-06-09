package campaignbus_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/zabolotny-dev/clicksafe/business/domain/campaignbus"
)

func TestParseCampaignType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    campaignbus.CampaignType
		wantErr bool
	}{
		{"EMAIL", campaignbus.EmailCampaign, false},
		{"MAX", campaignbus.MaxCampaign, false},
		{"UNKNOWN", campaignbus.CampaignType{}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := campaignbus.ParseCampaignType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCampaignType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseCampaignType(%q) = %v, want %v", tt.input, got, tt.want)
			}
			if !tt.wantErr && got.String() != tt.input {
				t.Errorf("String() = %v, want %v", got.String(), tt.input)
			}
		})
	}
}

func TestParseCampaignStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    campaignbus.CampaignStatus
		wantErr bool
	}{
		{"DRAFT", campaignbus.Draft, false},
		{"ACTIVE", campaignbus.Active, false},
		{"PAUSED", campaignbus.Paused, false},
		{"COMPLETED", campaignbus.Completed, false},
		{"CANCELED", campaignbus.Canceled, false},
		{"UNKNOWN", campaignbus.CampaignStatus{}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := campaignbus.ParseCampaignStatus(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCampaignStatus(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseCampaignStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
			if !tt.wantErr && got.String() != tt.input {
				t.Errorf("String() = %v, want %v", got.String(), tt.input)
			}
		})
	}
}

func TestParseTargetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		want    campaignbus.TargetStatus
		wantErr bool
	}{
		{"PENDING", campaignbus.Pending, false},
		{"SENT", campaignbus.Sent, false},
		{"FAILED", campaignbus.Failed, false},
		{"OPENED", campaignbus.Opened, false},
		{"CLICKED", campaignbus.Clicked, false},
		{"SUBMITTED", campaignbus.Submitted, false},
		{"REPLIED", campaignbus.Replied, false},
		{"UNKNOWN", campaignbus.TargetStatus{}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := campaignbus.ParseTargetStatus(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTargetStatus(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseTargetStatus(%q) = %v, want %v", tt.input, got, tt.want)
			}
			if !tt.wantErr && got.String() != tt.input {
				t.Errorf("String() = %v, want %v", got.String(), tt.input)
			}
		})
	}
}

func TestErrors_Message(t *testing.T) {
	t.Parallel()

	errUnscheduled := &campaignbus.ErrUnscheduledTargets{
		TargetIDs: []uuid.UUID{uuid.New(), uuid.New()},
	}
	if errUnscheduled.Error() != "targets without schedule: 2" {
		t.Errorf("ErrUnscheduledTargets.Error() = %v", errUnscheduled.Error())
	}

	errMissingVars := &campaignbus.ErrTargetsMissingVars{
		Targets: []campaignbus.TargetMissingVars{
			{TargetID: uuid.New()},
		},
	}
	if errMissingVars.Error() != "targets have missing required vars: 1" {
		t.Errorf("ErrTargetsMissingVars.Error() = %v", errMissingVars.Error())
	}
}
