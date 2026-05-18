package attachmentbus

import (
	"reflect"
	"testing"
)

func TestValidateAndExtractRequiredVarsAllowsCampaignDomain(t *testing.T) {
	content := []byte(`<img src="{{ .Campaign.Domain }}/attachment/example" alt="" />`)

	vars, err := validateAndExtractRequiredVars(content, Html)
	if err != nil {
		t.Fatalf("validateAndExtractRequiredVars returned error: %v", err)
	}

	expected := []string{"Campaign.Domain"}
	if !reflect.DeepEqual(vars, expected) {
		t.Fatalf("vars = %v, want %v", vars, expected)
	}
}
