package httpclient

import (
	"net/http"
	"time"
)

// NewClient creates an HTTP client configured for image validation (no redirects).
func NewClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		// Prevent following redirects for image validation
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
