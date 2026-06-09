package apitest

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/labstack/echo/v5"
)

// AddAuthCookie injects the session cookie and CSRF header into req.
// rawToken is the plain-text token returned by sessionbus.Create (goes into the cookie);
// csrfToken is the CSRF token (goes into X-CSRF-Token header).
func AddAuthCookie(req *http.Request, rawToken, csrfToken string) {
	cookie := &http.Cookie{
		Name:     "__Host-session",
		Value:    rawToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(8 * time.Hour),
	}
	req.AddCookie(cookie)
	req.Header.Set("X-CSRF-Token", csrfToken)
}

// PerformRequest executes an HTTP request on the given Echo engine and returns the recorder.
func PerformRequest(e *echo.Echo, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}
