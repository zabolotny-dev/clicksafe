package landingbus

import (
	"errors"
	"reflect"
	"testing"
)

func TestValidateAndExtractRequiredVarsRejectsCampaignRoot(t *testing.T) {
	t.Parallel()

	_, err := validateAndExtractRequiredVars([]byte("<p>{{ .Campaign.Label }}</p>"))
	if !errors.Is(err, ErrUnsupportedTemplateSyntax) {
		t.Fatalf("validateAndExtractRequiredVars error = %v, want %v", err, ErrUnsupportedTemplateSyntax)
	}
}

func TestValidateAndExtractRequiredVarsAllowsSupportedRoots(t *testing.T) {
	t.Parallel()

	vars, err := validateAndExtractRequiredVars([]byte("<p>{{ .Employee.FirstName }}</p><p>{{ .Target.Link }}</p>"))
	if err != nil {
		t.Fatalf("validateAndExtractRequiredVars error = %v", err)
	}

	expected := []string{"Employee.FirstName", "Target.Link"}
	if !reflect.DeepEqual(vars, expected) {
		t.Fatalf("validateAndExtractRequiredVars vars = %v, want %v", vars, expected)
	}
}
