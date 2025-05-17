package validator

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/monchh/annict-slack-bot/usecase"
)

const imageCheckTimeout = 5 * time.Second

// httpImageValidator implements the ImageValidationService using HTTP HEAD requests.
type httpImageValidator struct {
	httpClient HTTPClient // Interface for the infrastructure HTTP client
}

// HTTPClient defines the methods needed from the infrastructure HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPImageValidator creates a new validator instance.
func NewHTTPImageValidator(client HTTPClient) usecase.ImageValidationService {
	return &httpImageValidator{httpClient: client}
}

// ValidateURL checks if the URL points to a valid, non-redirecting image.
func (v *httpImageValidator) ValidateURL(ctx context.Context, url string) (isValid bool, validatedURL string) {
	if url == "" {
		return false, ""
	}

	// Create a context with timeout specific to this validation
	reqCtx, cancel := context.WithTimeout(ctx, imageCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodHead, url, nil)
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to create HEAD request for image %s: %v", url, err))
		return false, ""
	}

	resp, err := v.httpClient.Do(req) // Use the injected client
	if err != nil {
		if !strings.Contains(err.Error(), "context deadline exceeded") {
			slog.Warn(fmt.Sprintf("HEAD request failed for image %s: %v", url, err))
		}
		return false, ""
	}
	defer resp.Body.Close()

	// Check for 2xx success status code. Redirects are handled by the infra client config.
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		contentType := resp.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "image/") {
			slog.Debug(fmt.Sprintf("Valid image URL %s (Status: %d, Type: %s)", url, resp.StatusCode, contentType))
			return true, url
		}
		slog.Debug(fmt.Sprintf("URL %s has non-image Content-Type: %s", url, contentType))
	} else {
		slog.Debug(fmt.Sprintf("Invalid status code %d for image URL %s", resp.StatusCode, url))
	}

	return false, ""
}
