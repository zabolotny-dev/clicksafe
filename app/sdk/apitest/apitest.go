package apitest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/zabolotny-dev/clicksafe/business/sdk/dbtest"
)

// Table contains the logic for an API test.
type Table struct {
	Name       string
	URL        string
	Method     string
	StatusCode int
	Token      string // raw session token for cookie (set via AddAuthCookie)
	CSRFToken  string // csrf token for X-CSRF-Token header
	Input      any
	GotResp    any
	ExpResp    any
	CmpFunc    func(got any, exp any) string
}

// Test contains functions for executing an api test.
type Test struct {
	DB  *dbtest.Database
	Mux *echo.Echo
}

// Run performs the actual test logic based on the table data.
func (at *Test) Run(t *testing.T, table []Table, testName string) {
	for _, tt := range table {
		f := func(t *testing.T) {
			r := httptest.NewRequest(tt.Method, tt.URL, nil)

			if tt.Input != nil {
				d, err := json.Marshal(tt.Input)
				if err != nil {
					t.Fatalf("Should be able to marshal the model : %s", err)
				}
				r = httptest.NewRequest(tt.Method, tt.URL, bytes.NewBuffer(d))
				r.Header.Set("Content-Type", "application/json")
			}

			if tt.Token != "" {
				AddAuthCookie(r, tt.Token, tt.CSRFToken)
			}

			w := httptest.NewRecorder()
			at.Mux.ServeHTTP(w, r)

			if w.Code != tt.StatusCode {
				t.Logf("Response Body: %s", w.Body.String())
				t.Fatalf("%s: Should receive a status code of %d for the response : %d", tt.Name, tt.StatusCode, w.Code)
			}

			if tt.StatusCode == http.StatusNoContent {
				return
			}

			if tt.GotResp != nil {
				if err := json.Unmarshal(w.Body.Bytes(), tt.GotResp); err != nil {
					t.Fatalf("Should be able to unmarshal the response : %s\nBody: %s", err, w.Body.String())
				}

				if tt.CmpFunc != nil {
					diff := tt.CmpFunc(tt.GotResp, tt.ExpResp)
					if diff != "" {
						t.Log("DIFF")
						t.Logf("%s", diff)
						t.Log("GOT")
						t.Logf("%#v", tt.GotResp)
						t.Log("EXP")
						t.Logf("%#v", tt.ExpResp)
						t.Fatalf("Should get the expected response")
					}
				}
			}
		}

		t.Run(testName+"-"+tt.Name, f)
	}
}
